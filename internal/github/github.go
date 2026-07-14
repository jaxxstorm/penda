package github

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
	DefaultAPIURL      = "https://api.github.com"
	githubAcceptHeader = "application/vnd.github+json"
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

type Alert = alert

type alertFetcher interface {
	listAlerts(context.Context, repository, string) ([]alert, error)
}

type AlertFetcher interface {
	ListAlerts(context.Context, Repository, string) ([]Alert, error)
}

type githubClient struct {
	baseURL string
	client  *http.Client
}

type Client = githubClient

func NewClient(baseURL string, client *http.Client) *Client {
	return &Client{baseURL: baseURL, client: client}
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

func (client *githubClient) ListAlerts(ctx context.Context, repo Repository, token string) ([]Alert, error) {
	return client.listAlerts(ctx, repo, token)
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
