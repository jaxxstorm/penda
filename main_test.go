package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfigUsesExplicitDirectoryAndToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "environment-token")
	dir := t.TempDir()
	config, err := parseConfig(
		[]string{"--dir", dir, "--token", "flag-token"},
		func() (string, error) { return "", errors.New("getwd should not be called") },
		os.Getenv,
	)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if config.Dir != dir {
		t.Errorf("Dir = %q, want %q", config.Dir, dir)
	}
	if config.Token != "flag-token" {
		t.Errorf("Token = %q, want explicit token", config.Token)
	}
}

func TestParseConfigAcceptsShortDirectoryOption(t *testing.T) {
	dir := t.TempDir()
	config, err := parseConfig(
		[]string{"-d", dir},
		func() (string, error) { return "", errors.New("getwd should not be called") },
		func(string) string { return "" },
	)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if config.Dir != dir {
		t.Errorf("Dir = %q, want %q", config.Dir, dir)
	}
}

func TestParseConfigUsesCurrentDirectoryAndEnvironmentToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "environment-token")
	dir := t.TempDir()
	config, err := parseConfig(
		nil,
		func() (string, error) { return dir, nil },
		os.Getenv,
	)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if config.Dir != dir {
		t.Errorf("Dir = %q, want %q", config.Dir, dir)
	}
	if config.Token != "environment-token" {
		t.Errorf("Token = %q, want environment token", config.Token)
	}
}

func TestParseConfigUsesEnvironmentTokenForEmptyFlag(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "environment-token")
	config, err := parseConfig(
		[]string{"--token", ""},
		func() (string, error) { return t.TempDir(), nil },
		os.Getenv,
	)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if config.Token != "environment-token" {
		t.Errorf("Token = %q, want environment token", config.Token)
	}
}

func TestValidateDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := validateDirectory(dir); err != nil {
		t.Errorf("validateDirectory(%q) error = %v", dir, err)
	}

	missing := filepath.Join(dir, "missing")
	if err := validateDirectory(missing); err == nil {
		t.Error("validateDirectory(missing) error = nil, want an error")
	}

	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateDirectory(file); err == nil {
		t.Error("validateDirectory(file) error = nil, want an error")
	}
}

func TestRunDoesNotWriteToken(t *testing.T) {
	const token = "super-secret-token"
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"--dir", token, "--token", token},
		&stdout,
		&stderr,
		func() (string, error) { return "", errors.New("getwd should not be called") },
		func(string) string { return "environment-token" },
	)
	if code == 0 {
		t.Fatal("run() exit code = 0, want non-zero")
	}
	if strings.Contains(stdout.String()+stderr.String(), token) {
		t.Errorf("run output contains token: %q", stdout.String()+stderr.String())
	}
}

func TestRunDoesNotWriteTokenOnSuccess(t *testing.T) {
	const token = "super-secret-token"
	var stdout, stderr bytes.Buffer
	code := runWithRuntime(
		[]string{"--dir", t.TempDir(), "--token", token},
		&stdout,
		&stderr,
		func() (string, error) { return "", errors.New("getwd should not be called") },
		func(string) string { return "environment-token" },
		testRuntime([]alert{{Number: 1, PackageEcosystem: "npm", PackageName: "example", ManifestPath: "package.json", FirstPatchedVersion: "1.2.3"}}),
	)
	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), token) {
		t.Errorf("run output contains token: %q", stdout.String()+stderr.String())
	}
	if !strings.Contains(stdout.String(), "Run complete") || !strings.Contains(stdout.String(), "1 open alerts | 1 planned updates") {
		t.Errorf("stdout = %q, want run summary", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Resolving GitHub repository") || !strings.Contains(stdout.String(), "1 planned updates from 1 open Dependabot alerts") {
		t.Errorf("stdout = %q, want status bars", stdout.String())
	}
}

func TestRunRequiresTokenBeforeGitHubAccess(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	var stdout, stderr bytes.Buffer
	code := runWithRuntime(
		[]string{"--dir", t.TempDir()},
		&stdout,
		&stderr,
		func() (string, error) { return "", errors.New("getwd should not be called") },
		os.Getenv,
		runtime{},
	)
	if code == 0 {
		t.Fatal("runWithRuntime() exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "GitHub token is required") {
		t.Errorf("stderr = %q, want missing token error", stderr.String())
	}
}

func TestRunRedactsTokenFromProviderFailure(t *testing.T) {
	const token = "super-secret-token"
	var stdout, stderr bytes.Buffer
	code := runWithRuntime(
		[]string{"--dir", t.TempDir(), "--token", token},
		&stdout,
		&stderr,
		func() (string, error) { return "", errors.New("getwd should not be called") },
		func(string) string { return "" },
		runtime{
			discoverRepository: func(context.Context, string) (repository, error) {
				return repository{Owner: "octo-org", Name: "example"}, nil
			},
			alerts: stubFetcher{alerts: []alert{{Number: 1}}},
			providers: []alertProvider{providerFunc(func(context.Context, string, []alert) error {
				return errors.New("failed with " + token)
			})},
		},
	)
	if code == 0 {
		t.Fatal("runWithRuntime() exit code = 0, want non-zero")
	}
	if strings.Contains(stdout.String()+stderr.String(), token) {
		t.Errorf("run output contains token: %q", stdout.String()+stderr.String())
	}
}

func TestRunProvidersContinuesAfterFailure(t *testing.T) {
	var calls int
	err := runProviders(context.Background(), t.TempDir(), []alert{{Number: 1}}, []alertProvider{
		providerFunc(func(context.Context, string, []alert) error {
			calls++
			return errors.New("provider failure")
		}),
		providerFunc(func(context.Context, string, []alert) error {
			calls++
			return nil
		}),
	})
	if err == nil {
		t.Fatal("runProviders() error = nil, want an error")
	}
	if calls != 2 {
		t.Errorf("provider calls = %d, want 2", calls)
	}
}

func TestRunProvidersInvokesEachProvider(t *testing.T) {
	alerts := []alert{{Number: 1}}
	dir := t.TempDir()
	var calls int
	err := runProviders(context.Background(), dir, alerts, []alertProvider{
		providerFunc(func(_ context.Context, receivedDir string, received []alert) error {
			calls++
			if receivedDir != dir {
				t.Errorf("first provider directory = %q, want %q", receivedDir, dir)
			}
			if len(received) != 1 || received[0].Number != 1 {
				t.Errorf("first provider received %#v, want %#v", received, alerts)
			}
			return nil
		}),
		providerFunc(func(_ context.Context, receivedDir string, received []alert) error {
			calls++
			if receivedDir != dir {
				t.Errorf("second provider directory = %q, want %q", receivedDir, dir)
			}
			if len(received) != 1 || received[0].Number != 1 {
				t.Errorf("second provider received %#v, want %#v", received, alerts)
			}
			return nil
		}),
	})
	if err != nil {
		t.Fatalf("runProviders() error = %v", err)
	}
	if calls != 2 {
		t.Errorf("provider calls = %d, want 2", calls)
	}
}

func TestRunDoesNotModifyTargetWithoutProviders(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	if err := os.WriteFile(marker, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runWithRuntime(
		[]string{"--dir", dir, "--token", "token"},
		&stdout,
		&stderr,
		func() (string, error) { return "", errors.New("getwd should not be called") },
		func(string) string { return "" },
		testRuntime([]alert{{Number: 1}}),
	)
	if code != 0 {
		t.Fatalf("runWithRuntime() exit code = %d, want 0; stderr = %q", code, stderr.String())
	}
	content, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "unchanged" {
		t.Errorf("marker = %q, want unchanged", content)
	}
}

func TestWriteErrorShowsSafeCommandDiagnostic(t *testing.T) {
	const token = "super-secret-token"
	var output bytes.Buffer
	writeError(&output, &commandError{
		dir:    "/workspace/project",
		name:   "uv",
		args:   []string{"lock", "--upgrade-package", "pip"},
		output: "No solution found\nregistry token: " + token,
		err:    errors.New("exit status 1"),
	}, token)

	got := output.String()
	if !strings.Contains(got, "Command failed: uv lock --upgrade-package pip") {
		t.Errorf("error output = %q, want command diagnostic", got)
	}
	if !strings.Contains(got, "No solution found") {
		t.Errorf("error output = %q, want native output", got)
	}
	if strings.Contains(got, token) {
		t.Errorf("error output contains token: %q", got)
	}
}

func TestTrimCommandOutputKeepsRecentLines(t *testing.T) {
	lines := make([]string, 14)
	for index := range lines {
		lines[index] = fmt.Sprintf("line %d", index)
	}
	output := trimCommandOutput(strings.Join(lines, "\n"))
	if !strings.HasPrefix(output, "...\nline 2") || !strings.HasSuffix(output, "line 13") {
		t.Errorf("trimCommandOutput() = %q", output)
	}
}

type stubFetcher struct {
	alerts []alert
	err    error
}

func (fetcher stubFetcher) listAlerts(context.Context, repository, string) ([]alert, error) {
	return fetcher.alerts, fetcher.err
}

type providerFunc func(context.Context, string, []alert) error

func (provider providerFunc) process(ctx context.Context, dir string, alerts []alert, _ ...alertReporter) error {
	return provider(ctx, dir, alerts)
}

func testRuntime(alerts []alert) runtime {
	return runtime{
		discoverRepository: func(context.Context, string) (repository, error) {
			return repository{Owner: "octo-org", Name: "example"}, nil
		},
		alerts: stubFetcher{alerts: alerts},
	}
}
