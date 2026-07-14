package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type providerOps struct {
	runCommand      func(context.Context, string, string, ...string) error
	readFile        func(string) ([]byte, error)
	writeFileAtomic func(string, []byte, fs.FileMode) error
	stat            func(string) (fs.FileInfo, error)
}

type pythonProvider struct{ ops providerOps }
type npmProvider struct{ ops providerOps }
type githubActionsProvider struct{ ops providerOps }

type commandError struct {
	dir    string
	name   string
	args   []string
	output string
	err    error
}

func (err *commandError) Error() string {
	return "native package command failed"
}

func (err *commandError) Unwrap() error {
	return err.err
}

func defaultProviderOps() providerOps {
	return providerOps{
		runCommand: func(ctx context.Context, dir, name string, args ...string) error {
			command := exec.CommandContext(ctx, name, args...)
			command.Dir = dir
			output, err := command.CombinedOutput()
			if err != nil {
				return &commandError{dir: dir, name: name, args: args, output: string(output), err: err}
			}
			return nil
		},
		readFile:        os.ReadFile,
		writeFileAtomic: writeFileAtomic,
		stat:            os.Stat,
	}
}

func builtinProviders() []alertProvider {
	ops := defaultProviderOps()
	return []alertProvider{pythonProvider{ops: ops}, npmProvider{ops: ops}, githubActionsProvider{ops: ops}}
}

func (provider pythonProvider) process(ctx context.Context, dir string, alerts []alert, reports ...alertReporter) error {
	report := firstReporter(reports)
	for index, alert := range alerts {
		if alert.PackageName == "" || alert.FirstPatchedVersion == "" || !isPythonEcosystem(alert.PackageEcosystem) {
			continue
		}
		switch filepath.Base(alert.ManifestPath) {
		case "Pipfile":
			if strings.EqualFold(alert.PackageEcosystem, "pipenv") {
				if err := provider.runPackageCommand(ctx, dir, alert, index, report, "pipenv", "install", alert.PackageName+"=="+alert.FirstPatchedVersion); err != nil {
					return err
				}
			}
		case "pyproject.toml":
			if strings.EqualFold(alert.PackageEcosystem, "poetry") {
				if err := provider.runPackageCommand(ctx, dir, alert, index, report, "poetry", "add", alert.PackageName+"@"+alert.FirstPatchedVersion); err != nil {
					return err
				}
			}
		case "poetry.lock":
			if err := provider.runPackageCommand(ctx, dir, alert, index, report, "poetry", "update", alert.PackageName); err != nil {
				return err
			}
		case "uv.lock":
			if err := provider.runPackageCommand(ctx, dir, alert, index, report, "uv", "lock", "--upgrade-package", alert.PackageName); err != nil {
				return err
			}
		default:
			if !isRequirementsFile(alert.ManifestPath) || !strings.EqualFold(alert.PackageEcosystem, "pip") {
				continue
			}
			if err := provider.updatePip(ctx, dir, alert, index, report); err != nil {
				return err
			}
		}
	}
	return nil
}

func (provider pythonProvider) updatePip(ctx context.Context, dir string, alert alert, index int, report alertReporter) error {
	manifest, err := resolveManifest(dir, alert.ManifestPath)
	if err != nil {
		return err
	}
	content, err := provider.ops.readFile(manifest)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return errors.New("could not read Python requirements manifest")
	}
	updated, changed := pinRequirement(content, alert.PackageName, alert.FirstPatchedVersion)
	if !changed {
		return nil
	}
	reportAlert(report, index, alert)
	info, err := provider.ops.stat(manifest)
	if err != nil {
		return errors.New("could not inspect Python requirements manifest")
	}
	if err := provider.ops.writeFileAtomic(manifest, updated, info.Mode()); err != nil {
		return errors.New("could not update Python requirements manifest")
	}
	return provider.ops.runCommand(ctx, filepath.Dir(manifest), "python", "-m", "pip", "install", "-r", filepath.Base(manifest))
}

func (provider pythonProvider) runPackageCommand(ctx context.Context, dir string, alert alert, index int, report alertReporter, command string, args ...string) error {
	manifest, err := resolveManifest(dir, alert.ManifestPath)
	if err != nil {
		return err
	}
	info, err := provider.ops.stat(manifest)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return errors.New("could not inspect Python manifest")
	}
	original, err := provider.ops.readFile(manifest)
	if err != nil {
		return errors.New("could not read Python manifest")
	}
	reportAlert(report, index, alert)
	if err := provider.ops.runCommand(ctx, filepath.Dir(manifest), command, args...); err != nil {
		if restoreErr := provider.ops.writeFileAtomic(manifest, original, info.Mode()); restoreErr != nil {
			return errors.New("native package command failed and Penda could not restore the Python manifest")
		}
		return err
	}
	return nil
}

func (provider npmProvider) process(ctx context.Context, dir string, alerts []alert, reports ...alertReporter) error {
	report := firstReporter(reports)
	seen := make(map[string]struct{})
	for index, alert := range alerts {
		if strings.ToLower(alert.PackageEcosystem) != "npm" || alert.PackageName == "" || alert.FirstPatchedVersion == "" || !isNPMManifest(alert.ManifestPath) {
			continue
		}
		manifest, err := resolveManifest(dir, alert.ManifestPath)
		if err != nil {
			return err
		}
		if _, err := provider.ops.stat(manifest); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return errors.New("could not inspect npm manifest")
		}
		if filepath.Base(manifest) == "package-lock.json" {
			if _, err := provider.ops.stat(filepath.Join(filepath.Dir(manifest), "package.json")); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return errors.New("could not inspect npm manifest")
			}
		}
		args := []string{"install", alert.PackageName + "@" + alert.FirstPatchedVersion}
		if strings.EqualFold(alert.Scope, "development") || strings.EqualFold(alert.Scope, "dev") {
			args = append(args, "--save-dev")
		}
		key := manifest + "\x00" + strings.Join(args, "\x00")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		reportAlert(report, index, alert)
		if err := provider.ops.runCommand(ctx, filepath.Dir(manifest), "npm", args...); err != nil {
			return err
		}
	}
	return nil
}

func (provider githubActionsProvider) process(_ context.Context, dir string, alerts []alert, reports ...alertReporter) error {
	report := firstReporter(reports)
	for index, alert := range alerts {
		if !isGitHubActionsEcosystem(alert.PackageEcosystem) || alert.PackageName == "" || alert.FirstPatchedVersion == "" || !isWorkflowFile(alert.ManifestPath) {
			continue
		}
		manifest, err := resolveManifest(dir, alert.ManifestPath)
		if err != nil {
			return err
		}
		content, err := provider.ops.readFile(manifest)
		if err != nil {
			return errors.New("could not read GitHub Actions workflow")
		}
		updated, changed := updateActionReference(content, alert.PackageName, alert.FirstPatchedVersion)
		if !changed {
			continue
		}
		reportAlert(report, index, alert)
		info, err := provider.ops.stat(manifest)
		if err != nil {
			return errors.New("could not inspect GitHub Actions workflow")
		}
		if err := provider.ops.writeFileAtomic(manifest, updated, info.Mode()); err != nil {
			return errors.New("could not update GitHub Actions workflow")
		}
	}
	return nil
}

func resolveManifest(dir, manifestPath string) (string, error) {
	clean := filepath.Clean(manifestPath)
	if manifestPath == "" || filepath.IsAbs(manifestPath) || !filepath.IsLocal(clean) {
		return "", errors.New("alert manifest path is outside the target directory")
	}
	root, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.New("could not resolve target directory")
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	target := filepath.Join(root, clean)
	if !pathWithin(root, target) {
		return "", errors.New("alert manifest path is outside the target directory")
	}
	if resolved, err := filepath.EvalSymlinks(target); err == nil && !pathWithin(root, resolved) {
		return "", errors.New("alert manifest path is outside the target directory")
	}
	return target, nil
}

func pathWithin(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func isRequirementsFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.HasPrefix(name, "requirements") && strings.HasSuffix(name, ".txt")
}

func isNPMManifest(path string) bool {
	name := filepath.Base(path)
	return name == "package.json" || name == "package-lock.json"
}

func isWorkflowFile(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return strings.HasPrefix(clean, ".github/workflows/") && (strings.HasSuffix(clean, ".yml") || strings.HasSuffix(clean, ".yaml"))
}

func isGitHubActionsEcosystem(ecosystem string) bool {
	normalized := strings.ReplaceAll(strings.ToLower(ecosystem), "-", "_")
	return normalized == "github_actions"
}

func isPythonEcosystem(ecosystem string) bool {
	switch strings.ToLower(ecosystem) {
	case "pip", "pipenv", "poetry", "uv":
		return true
	default:
		return false
	}
}

func pinRequirement(content []byte, packageName, version string) ([]byte, bool) {
	matcher := regexp.MustCompile(`(?i)^(\s*([a-z0-9_.-]+)(?:\[[^\]]+\])?\s*)(?:[<>=!~].*)?$`)
	lines := strings.Split(string(content), "\n")
	for index, line := range lines {
		body, comment, _ := strings.Cut(line, " #")
		if match := matcher.FindStringSubmatch(body); match != nil && normalizePythonPackageName(match[2]) == normalizePythonPackageName(packageName) {
			lines[index] = match[1] + "==" + version
			if comment != "" {
				lines[index] += " #" + comment
			}
			return []byte(strings.Join(lines, "\n")), true
		}
	}
	return content, false
}

func normalizePythonPackageName(name string) string {
	return strings.Map(func(character rune) rune {
		if character == '-' || character == '_' || character == '.' {
			return '-'
		}
		return character
	}, strings.ToLower(name))
}

func updateActionReference(content []byte, packageName, version string) ([]byte, bool) {
	matcher := regexp.MustCompile(`(?m)^(\s*(?:-\s*)?uses:\s*)(['"]?)` + regexp.QuoteMeta(packageName) + `@[^\s#'"]+(['"]?)`)
	updated := matcher.ReplaceAllStringFunc(string(content), func(match string) string {
		parts := matcher.FindStringSubmatch(match)
		return parts[1] + parts[2] + packageName + "@" + version + parts[3]
	})
	return []byte(updated), updated != string(content)
}

func writeFileAtomic(path string, content []byte, mode fs.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".penda-*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if err := file.Chmod(mode); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func reportAlert(report alertReporter, index int, alert alert) {
	if report != nil {
		report(index, alert)
	}
}

func firstReporter(reports []alertReporter) alertReporter {
	if len(reports) == 0 {
		return nil
	}
	return reports[0]
}
