package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spencer17x/wtx/internal/core"
	"github.com/spencer17x/wtx/internal/fsops"
)

func TestParseArgvRecognizesDryRunAndNonInteractive(t *testing.T) {
	t.Parallel()

	parsed, err := parseArgv([]string{
		"add",
		"feature/cli",
		"--dry-run",
		"--non-interactive",
		"--mode",
		"default",
	})
	if err != nil {
		t.Fatalf("parse argv: %v", err)
	}

	if parsed.command != "add" {
		t.Fatalf("command = %q, want %q", parsed.command, "add")
	}

	if !parsed.options.dryRun {
		t.Fatal("dryRun = false, want true")
	}

	if !parsed.options.nonInteractive {
		t.Fatal("nonInteractive = false, want true")
	}
}

func TestParseArgvRecognizesBatchAddCommand(t *testing.T) {
	t.Parallel()

	parsed, err := parseArgv([]string{
		"batch-add",
		"feature/one",
		"feature/two",
		"--new-branch",
		"--mode",
		"default",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("parse argv: %v", err)
	}

	if parsed.command != "batch-add" {
		t.Fatalf("command = %q, want %q", parsed.command, "batch-add")
	}

	if len(parsed.options.batchBranches) != 2 {
		t.Fatalf("batch branches = %#v, want 2 branches", parsed.options.batchBranches)
	}

	if parsed.options.batchBranches[0] != "feature/one" || parsed.options.batchBranches[1] != "feature/two" {
		t.Fatalf("batch branches = %#v, want feature/one + feature/two", parsed.options.batchBranches)
	}

	if parsed.options.branchMode != core.BranchModeNew {
		t.Fatalf("branch mode = %q, want %q", parsed.options.branchMode, core.BranchModeNew)
	}
}

func TestShouldUsePromptsHonorsNonInteractiveMode(t *testing.T) {
	t.Parallel()

	if shouldUsePrompts(true, parsedAddOptions{}) != true {
		t.Fatal("expected prompts when shell is interactive and no flags disable them")
	}

	if shouldUsePrompts(true, parsedAddOptions{yes: true}) {
		t.Fatal("expected --yes to disable prompts")
	}

	if shouldUsePrompts(true, parsedAddOptions{nonInteractive: true}) {
		t.Fatal("expected --non-interactive to disable prompts")
	}

	if shouldUsePrompts(false, parsedAddOptions{}) {
		t.Fatal("expected non-tty shell to disable prompts")
	}
}

func TestExecutePlanDryRunSkipsGitAndFilesystemMutation(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	gitCalled := false
	applyCalled := false

	err := executePlan(
		&output,
		parsedAddOptions{dryRun: true},
		"/repo",
		"feature-cli",
		"feature/cli",
		core.BranchModeNew,
		core.GitCommand{
			Command: "git",
			Args:    []string{"worktree", "add", "/tmp/wt", "feature/cli"},
		},
		fsops.ApplyInitializationPlanInput{
			SourceRoot:        "/repo",
			WorktreeDirectory: "/tmp/wt",
		},
		map[string][]core.CommandHook{
			"beforeCreate": {
				{
					ID:          "before",
					Description: "Before hook",
					Command:     "echo",
					Args:        []string{"before"},
				},
			},
		},
		func(string, string, ...string) error {
			gitCalled = true
			return nil
		},
		func(fsops.ApplyInitializationPlanInput) error {
			applyCalled = true
			return nil
		},
		func(string, map[string]string, string, ...string) error {
			t.Fatal("expected dry-run to skip hook execution")
			return nil
		},
	)
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}

	if gitCalled {
		t.Fatal("expected dry-run to skip git execution")
	}

	if applyCalled {
		t.Fatal("expected dry-run to skip filesystem initialization")
	}

	if !strings.Contains(output.String(), "Dry run:") {
		t.Fatalf("output = %q, want dry run banner", output.String())
	}

	if !strings.Contains(output.String(), "No changes were made.") {
		t.Fatalf("output = %q, want no changes message", output.String())
	}
}
