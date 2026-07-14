package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type commandCall struct {
	dir  string
	name string
	args []string
}

func TestBuiltinProviders(t *testing.T) {
	providers := defaultRuntime().providers
	if len(providers) != 3 {
		t.Fatalf("providers = %d, want 3", len(providers))
	}
	if _, ok := providers[0].(pythonProvider); !ok {
		t.Errorf("first provider = %T, want pythonProvider", providers[0])
	}
	if _, ok := providers[1].(npmProvider); !ok {
		t.Errorf("second provider = %T, want npmProvider", providers[1])
	}
	if _, ok := providers[2].(githubActionsProvider); !ok {
		t.Errorf("third provider = %T, want githubActionsProvider", providers[2])
	}
}

func TestPinRequirementNormalizesPythonPackageNames(t *testing.T) {
	updated, changed := pinRequirement([]byte("python_dateutil>=2.0\n"), "python-dateutil", "3.0")
	if !changed {
		t.Fatal("pinRequirement() changed = false, want true")
	}
	if string(updated) != "python_dateutil==3.0\n" {
		t.Errorf("pinRequirement() = %q", updated)
	}
}

func TestPythonProviderUpdatesRequirements(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "requirements.txt")
	writeFixture(t, manifest, "example>=1.0 # keep\nother==1.0\n")
	var calls []commandCall
	provider := pythonProvider{ops: testProviderOps(&calls)}

	err := provider.process(context.Background(), dir, []alert{{
		PackageEcosystem:    "pip",
		PackageName:         "example",
		ManifestPath:        "requirements.txt",
		FirstPatchedVersion: "1.2.3",
	}})
	if err != nil {
		t.Fatalf("process() error = %v", err)
	}
	content := readFixture(t, manifest)
	if content != "example==1.2.3 # keep\nother==1.0\n" {
		t.Errorf("requirements = %q", content)
	}
	want := []commandCall{{dir: resolvedDir(t, dir), name: "python", args: []string{"-m", "pip", "install", "-r", "requirements.txt"}}}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestPythonProviderRunsPipenvAndPoetry(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "Pipfile"), "")
	writeFixture(t, filepath.Join(dir, "pyproject.toml"), "")
	var calls []commandCall
	provider := pythonProvider{ops: testProviderOps(&calls)}

	err := provider.process(context.Background(), dir, []alert{
		{PackageEcosystem: "pipenv", PackageName: "example", ManifestPath: "Pipfile", FirstPatchedVersion: "1.2.3"},
		{PackageEcosystem: "poetry", PackageName: "other", ManifestPath: "pyproject.toml", FirstPatchedVersion: "4.5.6"},
	})
	if err != nil {
		t.Fatalf("process() error = %v", err)
	}
	want := []commandCall{
		{dir: resolvedDir(t, dir), name: "pipenv", args: []string{"install", "example==1.2.3"}},
		{dir: resolvedDir(t, dir), name: "poetry", args: []string{"add", "other@4.5.6"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestPythonProviderHandlesPipClassifiedLockfiles(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "poetry.lock"), "")
	writeFixture(t, filepath.Join(dir, "uv.lock"), "")
	var calls []commandCall
	provider := pythonProvider{ops: testProviderOps(&calls)}

	err := provider.process(context.Background(), dir, []alert{
		{PackageEcosystem: "pip", PackageName: "requests", ManifestPath: "poetry.lock", FirstPatchedVersion: "2.32.4"},
		{PackageEcosystem: "pip", PackageName: "idna", ManifestPath: "uv.lock", FirstPatchedVersion: "3.15"},
	})
	if err != nil {
		t.Fatalf("process() error = %v", err)
	}
	want := []commandCall{
		{dir: resolvedDir(t, dir), name: "poetry", args: []string{"update", "requests"}},
		{dir: resolvedDir(t, dir), name: "uv", args: []string{"lock", "--upgrade-package", "idna"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestPythonProviderSkipsMissingManifest(t *testing.T) {
	var calls []commandCall
	provider := pythonProvider{ops: testProviderOps(&calls)}
	err := provider.process(context.Background(), t.TempDir(), []alert{
		{PackageEcosystem: "pip", PackageName: "idna", ManifestPath: "requirements.txt", FirstPatchedVersion: "3.15"},
		{PackageEcosystem: "pip", PackageName: "requests", ManifestPath: "uv.lock", FirstPatchedVersion: "2.32.4"},
	})
	if err != nil {
		t.Fatalf("process() error = %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("commands = %#v, want none", calls)
	}
}

func TestPythonProviderRestoresLockfileAfterCommandFailure(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "uv.lock")
	writeFixture(t, manifest, "original\n")
	ops := defaultProviderOps()
	ops.runCommand = func(context.Context, string, string, ...string) error {
		if err := os.WriteFile(manifest, []byte("changed\n"), 0o644); err != nil {
			return err
		}
		return errors.New("uv failed")
	}
	provider := pythonProvider{ops: ops}
	if err := provider.process(context.Background(), dir, []alert{{PackageEcosystem: "pip", PackageName: "pip", ManifestPath: "uv.lock", FirstPatchedVersion: "26.1.2"}}); err == nil {
		t.Fatal("process() error = nil, want command failure")
	}
	if got := readFixture(t, manifest); got != "original\n" {
		t.Errorf("uv.lock = %q, want restored content", got)
	}
}

func TestNpmProviderUsesTargetedCommands(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "package.json"), "{}")
	var calls []commandCall
	provider := npmProvider{ops: testProviderOps(&calls)}

	err := provider.process(context.Background(), dir, []alert{
		{PackageEcosystem: "npm", PackageName: "example", ManifestPath: "package.json", FirstPatchedVersion: "1.2.3"},
		{PackageEcosystem: "npm", PackageName: "dev-example", ManifestPath: "package.json", Scope: "development", FirstPatchedVersion: "4.5.6"},
		{PackageEcosystem: "npm", PackageName: "skip", ManifestPath: "package.json"},
	})
	if err != nil {
		t.Fatalf("process() error = %v", err)
	}
	want := []commandCall{
		{dir: resolvedDir(t, dir), name: "npm", args: []string{"install", "example@1.2.3"}},
		{dir: resolvedDir(t, dir), name: "npm", args: []string{"install", "dev-example@4.5.6", "--save-dev"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestNpmProviderSupportsAndDeduplicatesLockfileAlerts(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "package.json"), "{}")
	writeFixture(t, filepath.Join(dir, "package-lock.json"), "{}")
	var calls []commandCall
	provider := npmProvider{ops: testProviderOps(&calls)}
	npmAlert := alert{PackageEcosystem: "npm", PackageName: "undici", ManifestPath: "package-lock.json", Scope: "development", FirstPatchedVersion: "6.27.0"}
	if err := provider.process(context.Background(), dir, []alert{npmAlert, npmAlert}); err != nil {
		t.Fatalf("process() error = %v", err)
	}
	want := []commandCall{{dir: resolvedDir(t, dir), name: "npm", args: []string{"install", "undici@6.27.0", "--save-dev"}}}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestGitHubActionsProviderUpdatesOnlyMatchingReference(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, ".github", "workflows", "ci.yml")
	writeFixture(t, manifest, "steps:\n  - uses: actions/checkout@v3\n  - uses: actions/setup-node@v4\n")
	provider := githubActionsProvider{ops: defaultProviderOps()}

	err := provider.process(context.Background(), dir, []alert{{
		PackageEcosystem:    "github_actions",
		PackageName:         "actions/checkout",
		ManifestPath:        ".github/workflows/ci.yml",
		FirstPatchedVersion: "v4.1.0",
	}})
	if err != nil {
		t.Fatalf("process() error = %v", err)
	}
	content := readFixture(t, manifest)
	if !strings.Contains(content, "uses: actions/checkout@v4.1.0") {
		t.Errorf("workflow did not update checkout: %q", content)
	}
	if !strings.Contains(content, "uses: actions/setup-node@v4") {
		t.Errorf("workflow changed setup-node: %q", content)
	}
}

func TestProvidersRejectTraversalAndIgnoreUnsupportedAlerts(t *testing.T) {
	dir := t.TempDir()
	var calls []commandCall
	provider := npmProvider{ops: testProviderOps(&calls)}
	err := provider.process(context.Background(), dir, []alert{{
		PackageEcosystem:    "npm",
		PackageName:         "example",
		ManifestPath:        "../package.json",
		FirstPatchedVersion: "1.2.3",
	}})
	if err == nil {
		t.Fatal("process() error = nil, want traversal error")
	}
	if len(calls) != 0 {
		t.Errorf("commands = %#v, want none", calls)
	}

	provider = npmProvider{ops: testProviderOps(&calls)}
	if err := provider.process(context.Background(), dir, []alert{{PackageEcosystem: "pip", PackageName: "example", ManifestPath: "package.json", FirstPatchedVersion: "1.2.3"}}); err != nil {
		t.Fatalf("process() error = %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("commands = %#v, want none", calls)
	}
}

func TestProviderErrorStopsHandoff(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "package.json"), "{}")
	ops := testProviderOps(nil)
	ops.runCommand = func(context.Context, string, string, ...string) error { return errors.New("npm failed") }
	provider := npmProvider{ops: ops}
	if err := runProviders(context.Background(), dir, []alert{{PackageEcosystem: "npm", PackageName: "example", ManifestPath: "package.json", FirstPatchedVersion: "1.2.3"}}, []alertProvider{provider}); err == nil {
		t.Fatal("runProviders() error = nil, want provider error")
	}
}

func TestRunProvidersReportsCurrentAlertAndRemainingCount(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "requirements.txt"), "example==1.0\n")
	var calls []commandCall
	provider := pythonProvider{ops: testProviderOps(&calls)}
	alerts := []alert{
		{PackageEcosystem: "npm", PackageName: "ignored", ManifestPath: "package.json", FirstPatchedVersion: "1.0.0"},
		{PackageEcosystem: "pip", PackageName: "example", ManifestPath: "requirements.txt", FirstPatchedVersion: "1.2.3"},
	}
	var current, total int
	var reported alert
	err := runProviders(context.Background(), dir, alerts, []alertProvider{provider}, func(index, count int, alert alert) {
		current, total, reported = index, count, alert
	})
	if err != nil {
		t.Fatalf("runProviders() error = %v", err)
	}
	if current != 2 || total != 2 || reported.PackageName != "example" {
		t.Errorf("progress = %d/%d %#v", current, total, reported)
	}
}

func TestWriteAlertProgress(t *testing.T) {
	var output bytes.Buffer
	writeAlertProgress(&output)(2, 5, alert{PackageEcosystem: "pip", PackageName: "idna", FirstPatchedVersion: "3.15", ManifestPath: "requirements.txt"})
	if got := output.String(); !strings.Contains(got, "2/5 planned updates | 3 remaining") || !strings.Contains(got, "pip: idna -> 3.15") || !strings.Contains(got, "requirements.txt") {
		t.Errorf("progress output = %q", got)
	}
}

func TestUpdatePlanDeduplicatesAlertTargets(t *testing.T) {
	alerts := []alert{
		{PackageEcosystem: "npm", PackageName: "undici", ManifestPath: "package-lock.json", Scope: "development", FirstPatchedVersion: "6.27.0"},
		{PackageEcosystem: "npm", PackageName: "undici", ManifestPath: "package-lock.json", Scope: "development", FirstPatchedVersion: "6.27.0"},
		{PackageEcosystem: "pip", PackageName: "pip", ManifestPath: "uv.lock", FirstPatchedVersion: "26.1.2"},
	}
	plan := updatePlan(alerts)
	if len(plan) != 2 {
		t.Fatalf("planned updates = %d, want 2", len(plan))
	}
	if plan[alertUpdateKey(alerts[0])] != 1 || plan[alertUpdateKey(alerts[2])] != 2 {
		t.Errorf("plan = %#v", plan)
	}
}

func testProviderOps(calls *[]commandCall) providerOps {
	ops := defaultProviderOps()
	ops.runCommand = func(_ context.Context, dir, name string, args ...string) error {
		if calls != nil {
			*calls = append(*calls, commandCall{dir: dir, name: name, args: append([]string(nil), args...)})
		}
		return nil
	}
	return ops
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFixture(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func resolvedDir(t *testing.T, dir string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}
