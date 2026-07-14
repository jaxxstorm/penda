package providers

import (
	"context"
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
