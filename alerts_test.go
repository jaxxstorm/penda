package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubClientListsAndMapsAllAlerts(t *testing.T) {
	var requests int
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		if request.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("Accept") != githubAcceptHeader {
			t.Errorf("Accept = %q", request.Header.Get("Accept"))
		}
		if request.URL.Query().Get("per_page") != "100" {
			t.Errorf("per_page = %q", request.URL.Query().Get("per_page"))
		}
		if request.URL.Query().Get("state") != "open" {
			t.Errorf("state = %q, want open", request.URL.Query().Get("state"))
		}

		switch request.URL.Query().Get("page") {
		case "":
			w.Header().Set("Link", "<"+server.URL+"/repos/octo-org/example/dependabot/alerts?per_page=100&page=2>; rel=\"next\"")
			_, _ = w.Write([]byte(`[{
				"number": 1,
				"state": "open",
				"dependency": {"package": {"name": "example", "ecosystem": "npm"}, "manifest_path": "go.mod", "scope": "runtime"},
				"security_advisory": {"ghsa_id": "GHSA-123", "cve_id": "CVE-2026-1", "severity": "high", "vulnerable_version_range": "< 1.2.3"},
				"security_vulnerability": {"first_patched_version": {"identifier": "1.2.3"}}
			}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"number": 2, "state": "open"}]`))
		default:
			t.Errorf("unexpected page %q", request.URL.Query().Get("page"))
		}
	}))
	defer server.Close()

	client := &githubClient{baseURL: server.URL, client: server.Client()}
	alerts, err := client.listAlerts(context.Background(), repository{Owner: "octo-org", Name: "example"}, "token")
	if err != nil {
		t.Fatalf("listAlerts() error = %v", err)
	}
	if requests != 2 {
		t.Errorf("requests = %d, want 2", requests)
	}
	if len(alerts) != 2 {
		t.Fatalf("alerts = %#v, want two alerts", alerts)
	}
	want := alert{
		Number:                 1,
		State:                  "open",
		PackageEcosystem:       "npm",
		PackageName:            "example",
		ManifestPath:           "go.mod",
		Scope:                  "runtime",
		GHSAID:                 "GHSA-123",
		CVEID:                  "CVE-2026-1",
		Severity:               "high",
		VulnerableVersionRange: "< 1.2.3",
		FirstPatchedVersion:    "1.2.3",
	}
	if alerts[0] != want {
		t.Errorf("first alert = %#v, want %#v", alerts[0], want)
	}
}

func TestGitHubClientHandlesEmptyAlertsAndMissingToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := &githubClient{baseURL: server.URL, client: server.Client()}
	alerts, err := client.listAlerts(context.Background(), repository{Owner: "octo-org", Name: "example"}, "token")
	if err != nil {
		t.Fatalf("listAlerts() error = %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("alerts = %#v, want none", alerts)
	}

	if _, err := client.listAlerts(context.Background(), repository{}, ""); err == nil {
		t.Fatal("listAlerts() with empty token error = nil, want an error")
	}
}

func TestGitHubClientDoesNotExposeResponseBody(t *testing.T) {
	const responseBody = "private GitHub response"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client := &githubClient{baseURL: server.URL, client: server.Client()}
	_, err := client.listAlerts(context.Background(), repository{Owner: "octo-org", Name: "example"}, "token")
	if err == nil {
		t.Fatal("listAlerts() error = nil, want an error")
	}
	if strings.Contains(err.Error(), responseBody) {
		t.Errorf("error exposes response body: %q", err)
	}
}
