package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultGitHubAPIURL = "https://api.github.com"
	githubAcceptHeader  = "application/vnd.github+json"
)

type alert struct {
	Number                 int
	State                  string
	PackageEcosystem       string
	PackageName            string
	ManifestPath           string
	Scope                  string
	GHSAID                 string
	CVEID                  string
	Severity               string
	VulnerableVersionRange string
	FirstPatchedVersion    string
}

type alertFetcher interface {
	listAlerts(context.Context, repository, string) ([]alert, error)
}

type alertProvider interface {
	process(context.Context, string, []alert, ...alertReporter) error
}

type alertReporter func(int, alert)

type githubClient struct {
	baseURL string
	client  *http.Client
}

type githubAlert struct {
	Number     int    `json:"number"`
	State      string `json:"state"`
	Dependency struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		ManifestPath string `json:"manifest_path"`
		Scope        string `json:"scope"`
	} `json:"dependency"`
	SecurityAdvisory struct {
		GHSAID                 string `json:"ghsa_id"`
		CVEID                  string `json:"cve_id"`
		Severity               string `json:"severity"`
		VulnerableVersionRange string `json:"vulnerable_version_range"`
	} `json:"security_advisory"`
	SecurityVulnerability struct {
		FirstPatchedVersion *struct {
			Identifier string `json:"identifier"`
		} `json:"first_patched_version"`
	} `json:"security_vulnerability"`
}

func (client *githubClient) listAlerts(ctx context.Context, repo repository, token string) ([]alert, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("GitHub token is required")
	}

	next, err := url.Parse(strings.TrimRight(client.baseURL, "/") + "/repos/" + url.PathEscape(repo.Owner) + "/" + url.PathEscape(repo.Name) + "/dependabot/alerts")
	if err != nil {
		return nil, errors.New("could not construct the GitHub Dependabot alerts request")
	}
	query := next.Query()
	query.Set("per_page", "100")
	query.Set("state", "open")
	next.RawQuery = query.Encode()

	var alerts []alert
	seen := make(map[string]struct{})
	for next != nil {
		if _, ok := seen[next.String()]; ok {
			return nil, errors.New("GitHub Dependabot alerts pagination repeated a page")
		}
		seen[next.String()] = struct{}{}

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, next.String(), nil)
		if err != nil {
			return nil, errors.New("could not construct the GitHub Dependabot alerts request")
		}
		request.Header.Set("Accept", githubAcceptHeader)
		request.Header.Set("Authorization", "Bearer "+token)

		response, err := client.client.Do(request)
		if err != nil {
			return nil, errors.New("could not request GitHub Dependabot alerts")
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			response.Body.Close()
			return nil, fmt.Errorf("GitHub Dependabot alerts request failed with status %s", response.Status)
		}

		var page []githubAlert
		decodeErr := json.NewDecoder(response.Body).Decode(&page)
		response.Body.Close()
		if decodeErr != nil {
			return nil, errors.New("could not decode GitHub Dependabot alerts response")
		}
		for _, item := range page {
			alerts = append(alerts, mapAlert(item))
		}

		next, err = nextPage(response.Header.Get("Link"), next)
		if err != nil {
			return nil, err
		}
		if next != nil {
			query := next.Query()
			query.Set("state", "open")
			next.RawQuery = query.Encode()
		}
	}

	return alerts, nil
}

func mapAlert(item githubAlert) alert {
	mapped := alert{
		Number:                 item.Number,
		State:                  item.State,
		PackageEcosystem:       item.Dependency.Package.Ecosystem,
		PackageName:            item.Dependency.Package.Name,
		ManifestPath:           item.Dependency.ManifestPath,
		Scope:                  item.Dependency.Scope,
		GHSAID:                 item.SecurityAdvisory.GHSAID,
		CVEID:                  item.SecurityAdvisory.CVEID,
		Severity:               item.SecurityAdvisory.Severity,
		VulnerableVersionRange: item.SecurityAdvisory.VulnerableVersionRange,
	}
	if item.SecurityVulnerability.FirstPatchedVersion != nil {
		mapped.FirstPatchedVersion = item.SecurityVulnerability.FirstPatchedVersion.Identifier
	}
	return mapped
}

func nextPage(header string, current *url.URL) (*url.URL, error) {
	for _, part := range strings.Split(header, ",") {
		if !strings.Contains(part, "rel=\"next\"") && !strings.Contains(part, "rel=next") {
			continue
		}

		start := strings.Index(part, "<")
		end := strings.Index(part, ">")
		if start == -1 || end <= start+1 {
			return nil, errors.New("GitHub Dependabot alerts pagination link is invalid")
		}

		next, err := url.Parse(part[start+1 : end])
		if err != nil {
			return nil, errors.New("GitHub Dependabot alerts pagination link is invalid")
		}
		next = current.ResolveReference(next)
		if next.Host != current.Host || next.Scheme != current.Scheme {
			return nil, errors.New("GitHub Dependabot alerts pagination link is invalid")
		}
		return next, nil
	}

	return nil, nil
}

func runProviders(ctx context.Context, dir string, alerts []alert, providers []alertProvider, progress ...func(int, int, alert)) error {
	var providerErrors []error
	plan := updatePlan(alerts)
	for _, provider := range providers {
		report := func(index int, alert alert) {
			if len(progress) > 0 && progress[0] != nil {
				if current, ok := plan[alertUpdateKey(alert)]; ok {
					progress[0](current, len(plan), alert)
				}
			}
		}
		if err := provider.process(ctx, dir, alerts, report); err != nil {
			providerErrors = append(providerErrors, fmt.Errorf("process Dependabot alerts: %w", err))
		}
	}
	return errors.Join(providerErrors...)
}

func updatePlan(alerts []alert) map[string]int {
	plan := make(map[string]int)
	for _, alert := range alerts {
		if alert.PackageEcosystem == "" || alert.PackageName == "" || alert.ManifestPath == "" || alert.FirstPatchedVersion == "" {
			continue
		}
		key := alertUpdateKey(alert)
		if _, exists := plan[key]; !exists {
			plan[key] = len(plan) + 1
		}
	}
	return plan
}

func alertUpdateKey(alert alert) string {
	return strings.Join([]string{strings.ToLower(alert.PackageEcosystem), alert.ManifestPath, strings.ToLower(alert.PackageName), alert.FirstPatchedVersion, strings.ToLower(alert.Scope)}, "\x00")
}
