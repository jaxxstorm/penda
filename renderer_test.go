package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRendererRendersLifecycleCards(t *testing.T) {
	var stdout, stderr bytes.Buffer
	renderer := newRenderer(&stdout, &stderr)
	renderer.stage("Resolving GitHub repository")
	renderer.planning(6, 21)
	summary := newUpdateSummary(6)
	for _, current := range []int{1, 2, 3, 4, 5, 6} {
		summary.report(current, alert{PackageEcosystem: "npm", PackageName: "undici", FirstPatchedVersion: "6.27.0", ManifestPath: "package-lock.json", Scope: "development"})
	}
	renderer.summary(21, summary, nil)

	output := stdout.String()
	for _, text := range []string{"Penda", "Resolving GitHub repository", "Update plan", "6 planned updates from 21 open Dependabot alerts", "Run complete", "6 completed | 0 failed | 0 not attempted", "Applied updates", "npm: undici -> 6.27.0 via npm install --save-dev | package-lock.json"} {
		if !strings.Contains(output, text) {
			t.Errorf("output missing %q: %q", text, output)
		}
	}
}

func TestRendererRendersBubbleProgressWithoutInteractiveTerminal(t *testing.T) {
	var stdout, stderr bytes.Buffer
	renderer := newRenderer(&stdout, &stderr)
	renderer.update(2, 6, alert{PackageEcosystem: "npm", PackageName: "undici", FirstPatchedVersion: "6.27.0", ManifestPath: "package-lock.json"})

	output := stdout.String()
	for _, text := range []string{"Applying update", "2/6 planned updates | 4 remaining", "npm: undici -> 6.27.0", "package-lock.json", "█"} {
		if !strings.Contains(output, text) {
			t.Errorf("output missing %q: %q", text, output)
		}
	}
	if strings.Contains(output, "\x1b[?1049h") {
		t.Errorf("renderer entered an alternate screen: %q", output)
	}
}

func TestRendererRedactsFailureCredentials(t *testing.T) {
	const token = "super-secret-token"
	var stdout, stderr bytes.Buffer
	renderer := newRenderer(&stdout, &stderr)
	renderer.failure(&commandError{name: "uv", args: []string{"lock"}, output: "failed with " + token, err: errors.New("exit status 1")}, token)

	output := stderr.String()
	if !strings.Contains(output, "Update failed") || !strings.Contains(output, "Command failed: uv lock") {
		t.Errorf("failure output = %q", output)
	}
	if strings.Contains(output, token) {
		t.Errorf("failure output contains token: %q", output)
	}
}

func TestRendererReplacesLiveTicketWithoutInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	renderer := newRenderer(&stdout, &stderr)
	renderer.live = true
	renderer.stage("Resolving GitHub repository")
	renderer.update(1, 2, alert{PackageEcosystem: "npm", PackageName: "undici", FirstPatchedVersion: "6.27.0", ManifestPath: "package-lock.json"})

	output := stdout.String()
	if !strings.Contains(output, "\x1b[2K") {
		t.Errorf("live output did not clear the previous ticket: %q", output)
	}
	if strings.Contains(output, "\x1b[?1049h") {
		t.Errorf("live output entered an alternate screen: %q", output)
	}
}

func TestRendererSummarizesFailures(t *testing.T) {
	var stdout, stderr bytes.Buffer
	renderer := newRenderer(&stdout, &stderr)
	summary := newUpdateSummary(6)
	for _, current := range []int{1, 2, 3, 4, 5, 6} {
		summary.report(current, alert{PackageEcosystem: "pip", PackageName: "pip", FirstPatchedVersion: "26.1.2", ManifestPath: "uv.lock"})
	}
	renderer.summary(21, summary, errors.New("native command failed"))

	output := stdout.String()
	for _, text := range []string{"Run completed with errors", "5 completed | 1 failed | 0 not attempted", "native command failed"} {
		if !strings.Contains(output, text) {
			t.Errorf("summary missing %q: %q", text, output)
		}
	}
}

func TestUpdateMethodDescribesNativeTooling(t *testing.T) {
	tests := []struct {
		alert alert
		want  string
	}{
		{alert: alert{PackageEcosystem: "npm", Scope: "development"}, want: "npm install --save-dev"},
		{alert: alert{PackageEcosystem: "pip", ManifestPath: "uv.lock"}, want: "uv lock --upgrade-package"},
		{alert: alert{PackageEcosystem: "github_actions"}, want: "workflow uses: update"},
	}
	for _, test := range tests {
		if got := updateMethod(test.alert); got != test.want {
			t.Errorf("updateMethod(%#v) = %q, want %q", test.alert, got, test.want)
		}
	}
}
