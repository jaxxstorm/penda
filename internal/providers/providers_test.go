package providers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jaxxstorm/penda/internal/github"
)

func TestPlanDeduplicatesEquivalentUpdates(t *testing.T) {
	alerts := []github.Alert{
		{PackageEcosystem: "npm", PackageName: "undici", ManifestPath: "package-lock.json", Scope: "development", FirstPatchedVersion: "6.27.0"},
		{PackageEcosystem: "npm", PackageName: "undici", ManifestPath: "package-lock.json", Scope: "development", FirstPatchedVersion: "6.27.0"},
	}
	if plan := Plan(alerts); len(plan) != 1 {
		t.Errorf("planned updates = %d, want 1", len(plan))
	}
}

func TestResolvedInManifestRequiresPackageAndVersionOnAddedLine(t *testing.T) {
	alert := github.Alert{PackageName: "undici", FirstPatchedVersion: "6.27.0", ManifestPath: "package-lock.json"}
	ops := defaultProviderOps()
	ops.manifestDiff = func(string, string) ([]byte, error) {
		return []byte("+    \"undici\": \"6.27.0\"\n"), nil
	}
	if !resolvedInManifest(ops, "/workspace", alert) {
		t.Fatal("resolvedInManifest() = false, want true")
	}
	ops.manifestDiff = func(string, string) ([]byte, error) {
		return []byte("+    \"other\": \"6.27.0\"\n"), nil
	}
	if resolvedInManifest(ops, "/workspace", alert) {
		t.Fatal("resolvedInManifest() = true, want false")
	}
}

func TestNpmProviderSkipsResolvedManifestDiff(t *testing.T) {
	ops := defaultProviderOps()
	ops.manifestDiff = func(string, string) ([]byte, error) {
		return []byte("+ \"undici\": \"6.27.0\"\n"), nil
	}
	ops.runCommand = func(context.Context, string, string, ...string) error {
		t.Fatal("runCommand() called for an already resolved alert")
		return nil
	}
	provider := npmProvider{ops: ops}
	if err := provider.Process(context.Background(), t.TempDir(), []github.Alert{{PackageEcosystem: "npm", PackageName: "undici", ManifestPath: "package-lock.json", FirstPatchedVersion: "6.27.0"}}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestBuiltinIncludesGoModuleProvider(t *testing.T) {
	for _, provider := range Builtin() {
		if _, ok := provider.(goModuleProvider); ok {
			return
		}
	}
	t.Fatal("Builtin() did not include goModuleProvider")
}

func TestGoModuleProviderUpdatesNestedModuleOnce(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte("module example.com/api\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var commandCalls int
	var commandDir, command string
	var args []string
	ops := defaultProviderOps()
	ops.manifestDiff = nil
	ops.runCommand = func(_ context.Context, dir, name string, arguments ...string) error {
		commandCalls++
		commandDir, command, args = dir, name, arguments
		return nil
	}
	provider := goModuleProvider{ops: ops}
	alerts := []github.Alert{
		{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "services/api/go.mod", FirstPatchedVersion: "v1.2.3"},
		{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "services/api/go.mod", FirstPatchedVersion: "v1.2.3"},
	}
	if err := provider.Process(context.Background(), root, alerts); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if commandCalls != 1 {
		t.Errorf("runCommand() calls = %d, want 1", commandCalls)
	}
	resolvedModuleDir, err := filepath.EvalSymlinks(moduleDir)
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}
	if commandDir != resolvedModuleDir {
		t.Errorf("command directory = %q, want %q", commandDir, resolvedModuleDir)
	}
	if command != "go" {
		t.Errorf("command = %q, want go", command)
	}
	if len(args) != 2 || args[0] != "get" || args[1] != "example.com/module@v1.2.3" {
		t.Errorf("command arguments = %q, want [get example.com/module@v1.2.3]", args)
	}
}

func TestGoModuleProviderSkipsResolvedAlert(t *testing.T) {
	ops := defaultProviderOps()
	ops.manifestDiff = func(string, string) ([]byte, error) {
		return []byte("+require example.com/module v1.2.3\n"), nil
	}
	ops.runCommand = func(context.Context, string, string, ...string) error {
		t.Fatal("runCommand() called for an already resolved alert")
		return nil
	}
	provider := goModuleProvider{ops: ops}
	alert := github.Alert{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "go.mod", FirstPatchedVersion: "v1.2.3"}
	if err := provider.Process(context.Background(), t.TempDir(), []github.Alert{alert}); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestGoModuleProviderSkipsIncompleteOrUnsupportedAlerts(t *testing.T) {
	var commandCalls int
	ops := defaultProviderOps()
	ops.manifestDiff = nil
	ops.runCommand = func(context.Context, string, string, ...string) error {
		commandCalls++
		return nil
	}
	provider := goModuleProvider{ops: ops}
	alerts := []github.Alert{
		{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "go.mod"},
		{PackageEcosystem: "npm", PackageName: "example.com/module", ManifestPath: "go.mod", FirstPatchedVersion: "v1.2.3"},
		{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "go.sum", FirstPatchedVersion: "v1.2.3"},
		{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "missing/go.mod", FirstPatchedVersion: "v1.2.3"},
	}
	if err := provider.Process(context.Background(), t.TempDir(), alerts); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if commandCalls != 0 {
		t.Errorf("runCommand() calls = %d, want 0", commandCalls)
	}
}

func TestGoModuleProviderRejectsTraversalManifest(t *testing.T) {
	ops := defaultProviderOps()
	ops.manifestDiff = nil
	ops.runCommand = func(context.Context, string, string, ...string) error {
		t.Fatal("runCommand() called for a traversal manifest")
		return nil
	}
	provider := goModuleProvider{ops: ops}
	alert := github.Alert{PackageEcosystem: "gomod", PackageName: "example.com/module", ManifestPath: "../go.mod", FirstPatchedVersion: "v1.2.3"}
	if err := provider.Process(context.Background(), t.TempDir(), []github.Alert{alert}); err == nil {
		t.Fatal("Process() error = nil, want traversal path error")
	}
}
