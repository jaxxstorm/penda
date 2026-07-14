package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kong"
)

type command struct {
	Dir   string `help:"Directory to scan." short:"d" type:"path"`
	Token string `env:"GITHUB_TOKEN" help:"GitHub access token."`
}

type config struct {
	Dir   string
	Token string
}

type runtime struct {
	discoverRepository func(context.Context, string) (repository, error)
	alerts             alertFetcher
	providers          []alertProvider
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getwd, os.Getenv))
}

func run(args []string, stdout, stderr io.Writer, getwd func() (string, error), getenv func(string) string) int {
	return runWithRuntime(args, stdout, stderr, getwd, getenv, defaultRuntime())
}

func defaultRuntime() runtime {
	return runtime{
		discoverRepository: discoverRepository,
		alerts: &githubClient{
			baseURL: defaultGitHubAPIURL,
			client:  http.DefaultClient,
		},
		providers: builtinProviders(),
	}
}

func runWithRuntime(args []string, stdout, stderr io.Writer, getwd func() (string, error), getenv func(string) string, runtime runtime) int {
	renderer := newRenderer(stdout, stderr)
	config, err := parseConfig(args, getwd, getenv)
	if err != nil {
		renderer.failure(err)
		return 2
	}

	if err := validateDirectory(config.Dir); err != nil {
		renderer.failure(err, config.Token)
		return 1
	}

	if strings.TrimSpace(config.Token) == "" {
		renderer.failure(errors.New("GitHub token is required"))
		return 1
	}

	renderer.stage("Resolving GitHub repository")
	repository, err := runtime.discoverRepository(context.Background(), config.Dir)
	if err != nil {
		renderer.failure(err, config.Token)
		return 1
	}

	renderer.stage("Retrieving Dependabot alerts")
	alerts, err := runtime.alerts.listAlerts(context.Background(), repository, config.Token)
	if err != nil {
		renderer.failure(err, config.Token)
		return 1
	}

	summary := newUpdateSummary(len(updatePlan(alerts)))
	renderer.planning(summary.planned, len(alerts))
	progress := func(current, total int, alert alert) {
		summary.report(current, alert)
		renderer.update(current, total, alert)
	}
	if err := runProviders(context.Background(), config.Dir, alerts, runtime.providers, progress); err != nil {
		renderer.summary(len(alerts), summary, err, config.Token)
		return 1
	}

	renderer.summary(len(alerts), summary, nil)
	return 0
}

func parseConfig(args []string, getwd func() (string, error), getenv func(string) string) (config, error) {
	var command command
	parser, err := kong.New(&command, kong.Name("penda"))
	if err != nil {
		return config{}, err
	}

	if _, err := parser.Parse(args); err != nil {
		return config{}, err
	}

	dir := command.Dir
	if dir == "" {
		dir, err = getwd()
		if err != nil {
			return config{}, fmt.Errorf("get current directory: %w", err)
		}
	}

	token := command.Token
	if strings.TrimSpace(token) == "" {
		token = getenv("GITHUB_TOKEN")
	}

	return config{Dir: dir, Token: token}, nil
}

func validateDirectory(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.ErrNotExist
		}
		return errors.New("target directory cannot be accessed")
	}
	if !info.IsDir() {
		return errors.New("target path is not a directory")
	}
	return nil
}

func writeAlertStatus(w io.Writer, count int) {
	newRenderer(w, io.Discard).complete(count)
}

func writeStatus(w io.Writer, message string) {
	newRenderer(w, io.Discard).stage(message)
}

func writeAlertProgress(w io.Writer) func(int, int, alert) {
	renderer := newRenderer(w, io.Discard)
	return renderer.update
}

func writeError(w io.Writer, err error, secrets ...string) {
	newRenderer(io.Discard, w).failure(err, secrets...)
}

func safeError(err error) string {
	if err == nil {
		return "unknown error"
	}
	if errors.Is(err, os.ErrNotExist) {
		return "target directory does not exist"
	}
	var commandErr *commandError
	if errors.As(err, &commandErr) {
		command := strings.TrimSpace(commandErr.name + " " + strings.Join(commandErr.args, " "))
		message := fmt.Sprintf("Command failed: %s\nDirectory: %s", command, commandErr.dir)
		if output := trimCommandOutput(commandErr.output); output != "" {
			message += "\n\n" + output
		}
		return message
	}
	return err.Error()
}

func trimCommandOutput(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	const maxLines = 12
	if len(lines) > maxLines {
		lines = append([]string{"..."}, lines[len(lines)-maxLines:]...)
	}
	return strings.Join(lines, "\n")
}
