package github

import (
	"context"
	"errors"
	"net/url"
	"os/exec"
	"strings"
)

type repository struct {
	Owner string
	Name  string
}

type Repository = repository

func DiscoverRepository(ctx context.Context, dir string) (Repository, error) {
	return discoverRepository(ctx, dir)
}

func discoverRepository(ctx context.Context, dir string) (repository, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return repository{}, errors.New("could not read the origin Git remote")
	}

	return parseRepositoryURL(strings.TrimSpace(string(output)))
}

func parseRepositoryURL(remote string) (repository, error) {
	if strings.HasPrefix(remote, "git@github.com:") {
		remote = "ssh://" + strings.Replace(remote, ":", "/", 1)
	}

	parsed, err := url.Parse(remote)
	if err != nil || parsed.Hostname() != "github.com" {
		return repository{}, errors.New("origin remote is not a github.com repository")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "ssh" {
		return repository{}, errors.New("origin remote is not a supported GitHub URL")
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return repository{}, errors.New("origin remote does not identify a GitHub repository")
	}

	name := strings.TrimSuffix(parts[1], ".git")
	if name == "" {
		return repository{}, errors.New("origin remote does not identify a GitHub repository")
	}

	return repository{Owner: parts[0], Name: name}, nil
}
