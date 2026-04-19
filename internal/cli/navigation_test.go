package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseArgvRecognizesPathCommand(t *testing.T) {
	t.Parallel()

	pathParsed, err := parseArgv([]string{"path", "feature/demo"})
	if err != nil {
		t.Fatalf("parseArgv path: %v", err)
	}
	if pathParsed.command != "path" || pathParsed.target != "feature/demo" {
		t.Fatalf("pathParsed = %#v", pathParsed)
	}
}

func TestParseArgvRecognizesShellInitCommand(t *testing.T) {
	t.Parallel()

	parsed, err := parseArgv([]string{"shell-init", "zsh"})
	if err != nil {
		t.Fatalf("parseArgv shell-init: %v", err)
	}
	if parsed.command != "shell-init" || parsed.target != "zsh" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestRunPathPrintsResolvedAbsolutePath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := runPath(&stdout, "feature/demo", func(branchName string) (string, error) {
		if branchName != "feature/demo" {
			t.Fatalf("branchName = %q", branchName)
		}
		return "/tmp/feature-demo", nil
	})
	if err != nil {
		t.Fatalf("runPath: %v", err)
	}

	if got := stdout.String(); got != "/tmp/feature-demo\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunSwitchPrintsPathAndGuidance(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := runSwitch(&stdout, "feature/demo", func(branchName string) (string, error) {
		if branchName != "feature/demo" {
			t.Fatalf("branchName = %q", branchName)
		}
		return "/tmp/feature-demo", nil
	})
	if err != nil {
		t.Fatalf("runSwitch: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "/tmp/feature-demo") {
		t.Fatalf("output = %q", output)
	}
	if !strings.Contains(output, "cd \"$(wtx path feature/demo)\"") {
		t.Fatalf("output = %q", output)
	}
}

func TestRunPathReturnsResolverError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("no worktree found for branch: feature/demo")
	err := runPath(&bytes.Buffer{}, "feature/demo", func(string) (string, error) {
		return "", wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestParseArgvRecognizesSwitchCommand(t *testing.T) {
	t.Parallel()

	parsed, err := parseArgv([]string{"switch", "feature/demo"})
	if err != nil {
		t.Fatalf("parseArgv switch: %v", err)
	}
	if parsed.command != "switch" || parsed.target != "feature/demo" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestPrintUsageIncludesNavigationCommands(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	printUsage(&stdout)

	output := stdout.String()
	if !strings.Contains(output, "wtx path <branch>") {
		t.Fatalf("usage missing path command: %q", output)
	}
}
