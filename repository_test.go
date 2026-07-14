package main

import (
	"context"
	"os/exec"
	"testing"
)

func TestParseRepositoryURL(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		want   repository
		fails  bool
	}{
		{name: "HTTPS", remote: "https://github.com/octo-org/example.git", want: repository{Owner: "octo-org", Name: "example"}},
		{name: "SSH URL", remote: "ssh://git@github.com/octo-org/example.git", want: repository{Owner: "octo-org", Name: "example"}},
		{name: "SCP-style SSH", remote: "git@github.com:octo-org/example.git", want: repository{Owner: "octo-org", Name: "example"}},
		{name: "missing repository", remote: "https://github.com/octo-org", fails: true},
		{name: "non-GitHub host", remote: "https://gitlab.com/octo-org/example.git", fails: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseRepositoryURL(test.remote)
			if test.fails {
				if err == nil {
					t.Fatal("parseRepositoryURL() error = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRepositoryURL() error = %v", err)
			}
			if got != test.want {
				t.Errorf("parseRepositoryURL() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestDiscoverRepository(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", "git@github.com:octo-org/example.git")

	got, err := discoverRepository(context.Background(), dir)
	if err != nil {
		t.Fatalf("discoverRepository() error = %v", err)
	}
	want := repository{Owner: "octo-org", Name: "example"}
	if got != want {
		t.Errorf("discoverRepository() = %#v, want %#v", got, want)
	}
}

func TestDiscoverRepositoryWithoutOrigin(t *testing.T) {
	if _, err := discoverRepository(context.Background(), t.TempDir()); err == nil {
		t.Fatal("discoverRepository() error = nil, want an error")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
