package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func splitLines(value string) []string {
	lines := strings.Split(value, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func captureCommand(cwd string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}

		return "", fmt.Errorf("command failed (%s %s): %s", command, strings.Join(args, " "), message)
	}

	return stdout.String(), nil
}

func ResolveRepoRoot(cwd string) (string, error) {
	output, err := captureCommand(cwd, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func ResolveCurrentBranch(cwd string) (string, error) {
	output, err := captureCommand(cwd, "git", "branch", "--show-current")
	if err != nil {
		return "", err
	}

	branch := strings.TrimSpace(output)
	if branch == "" {
		return "HEAD", nil
	}

	return branch, nil
}

func ListIgnoredPaths(repoRoot string) ([]string, error) {
	output, err := captureCommand(repoRoot, "git", "ls-files", "--others", "-i", "--exclude-standard", "--directory")
	if err != nil {
		return nil, err
	}

	return splitLines(output), nil
}

func ListProjectFiles(repoRoot string) ([]string, error) {
	output, err := captureCommand(repoRoot, "git", "ls-files", "--cached", "--others", "--exclude-standard", "--full-name")
	if err != nil {
		return nil, err
	}

	return splitLines(output), nil
}
