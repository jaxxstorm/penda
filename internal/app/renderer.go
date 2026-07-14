package app

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

type renderer struct {
	stdout   io.Writer
	stderr   io.Writer
	progress progress.Model
	live     bool
	lines    int
}

type updateSummary struct {
	planned   int
	attempted map[int]alert
}

func newUpdateSummary(planned int) *updateSummary {
	return &updateSummary{planned: planned, attempted: make(map[int]alert)}
}

func (summary *updateSummary) report(current int, alert alert) {
	summary.attempted[current] = alert
}

func newRenderer(stdout, stderr io.Writer) *renderer {
	bar := progress.New(progress.WithWidth(30), progress.WithDefaultGradient())
	bar.ShowPercentage = false
	return &renderer{stdout: stdout, stderr: stderr, progress: bar, live: isTerminal(stdout)}
}

func (renderer *renderer) stage(message string) {
	renderer.writeTicket("Penda", message, lipgloss.Color("62"))
}

func (renderer *renderer) planning(updates, alerts int) {
	renderer.writeTicket("Update plan", fmt.Sprintf("%d planned updates from %d open Dependabot alerts", updates, alerts), lipgloss.Color("33"))
}

func (renderer *renderer) update(current, total int, alert alert) {
	if total == 0 {
		return
	}
	percent := float64(current) / float64(total)
	remaining := total - current
	body := strings.Join([]string{
		renderer.progress.ViewAs(percent),
		fmt.Sprintf("%d/%d planned updates | %d remaining", current, total, remaining),
		fmt.Sprintf("%s: %s -> %s", alert.PackageEcosystem, alert.PackageName, alert.FirstPatchedVersion),
		alert.ManifestPath,
	}, "\n")
	renderer.writeTicket("Applying update", body, lipgloss.Color("36"))
}

func (renderer *renderer) complete(alerts int) {
	renderer.writeTicket("Complete", fmt.Sprintf("Processed %d open Dependabot alerts", alerts), lipgloss.Color("42"))
}

func (renderer *renderer) summary(alerts int, summary *updateSummary, err error, secrets ...string) {
	failed := countErrors(err)
	attempted := len(summary.attempted)
	completed := max(attempted-failed, 0)
	notAttempted := max(summary.planned-attempted, 0)
	title := "Run complete"
	color := lipgloss.Color("42")
	if err != nil {
		title = "Run completed with errors"
		color = lipgloss.Color("160")
	}
	body := fmt.Sprintf("%d open alerts | %d planned updates\n%d completed | %d failed | %d skipped or not attempted", alerts, summary.planned, completed, failed, notAttempted)
	if attempted > 0 {
		label := "Applied updates"
		if err != nil {
			label = "Attempted updates"
		}
		body += "\n\n" + label + ":\n" + strings.Join(summary.details(), "\n")
	}
	if err != nil {
		body += "\n\n" + redactOutput(safeError(err), secrets...)
	}
	renderer.writeTicket(title, body, color)
}

func (summary *updateSummary) details() []string {
	indexes := make([]int, 0, len(summary.attempted))
	for index := range summary.attempted {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	details := make([]string, 0, len(indexes))
	for _, index := range indexes {
		alert := summary.attempted[index]
		details = append(details, fmt.Sprintf("- %s: %s -> %s via %s | %s", alert.PackageEcosystem, alert.PackageName, alert.FirstPatchedVersion, updateMethod(alert), alert.ManifestPath))
	}
	return details
}

func updateMethod(alert alert) string {
	switch strings.ToLower(alert.PackageEcosystem) {
	case "npm":
		if strings.EqualFold(alert.Scope, "development") || strings.EqualFold(alert.Scope, "dev") {
			return "npm install --save-dev"
		}
		return "npm install"
	case "github_actions", "github-actions":
		return "workflow uses: update"
	case "pip", "pipenv", "poetry", "uv":
		switch filepath.Base(alert.ManifestPath) {
		case "uv.lock":
			return "uv lock --upgrade-package"
		case "poetry.lock":
			return "poetry update"
		case "Pipfile":
			return "pipenv install"
		case "pyproject.toml":
			return "poetry add"
		default:
			return "python -m pip install -r"
		}
	default:
		return "native update"
	}
}

func (renderer *renderer) failure(err error, secrets ...string) {
	message := redactOutput(safeError(err), secrets...)
	if renderer.live {
		renderer.writeTicket("Update failed", message, lipgloss.Color("160"))
		return
	}
	renderer.writeCard(renderer.stderr, "Update failed", message, lipgloss.Color("160"))
}

func (renderer *renderer) writeTicket(title, body string, color lipgloss.Color) {
	card := renderer.card(title, body, color)
	if !renderer.live {
		fmt.Fprintln(renderer.stdout, card)
		return
	}
	if renderer.lines > 0 {
		fmt.Fprintf(renderer.stdout, "\x1b[%dA", renderer.lines)
		for range renderer.lines {
			fmt.Fprint(renderer.stdout, "\x1b[2K\x1b[1B")
		}
		fmt.Fprintf(renderer.stdout, "\x1b[%dA", renderer.lines)
	}
	fmt.Fprintln(renderer.stdout, card)
	renderer.lines = strings.Count(card, "\n") + 1
}

func (renderer *renderer) writeCard(writer io.Writer, title, body string, color lipgloss.Color) {
	fmt.Fprintln(writer, renderer.card(title, body, color))
}

func (renderer *renderer) card(title, body string, color lipgloss.Color) string {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(color).Padding(0, 1).Render(title)
	content := lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.NewStyle().Padding(0, 1).Render(body))
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(color).Render(content)
}

func isTerminal(writer io.Writer) bool {
	file, ok := writer.(interface{ Fd() uintptr })
	return ok && (isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd()))
}

func redactOutput(message string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[redacted]")
		}
	}
	return message
}

func countErrors(err error) int {
	if err == nil {
		return 0
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		count := 0
		for _, nested := range joined.Unwrap() {
			count += countErrors(nested)
		}
		return count
	}
	return 1
}
