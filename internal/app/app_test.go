package app

import (
	"errors"
	"testing"
)

func TestParseConfigUsesExplicitTokenAndDirectory(t *testing.T) {
	config, err := parseConfig(
		[]string{"--dir", "/workspace", "--token", "flag-token"},
		func() (string, error) { return "", errors.New("getwd should not be called") },
		func(string) string { return "environment-token" },
	)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if config.Dir != "/workspace" || config.Token != "flag-token" {
		t.Errorf("config = %#v", config)
	}
}
