package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spencer17x/wtx/internal/config"
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

func TestResolveBranchModeUsesSelectedInteractiveChoice(t *testing.T) {
	t.Parallel()

	var gotMessage string
	var gotChoices []string
	var gotDefaultIndex int
	ui := &promptUI{
		selectPrompter: func(message string, choices []string, defaultIndex int) (string, error) {
			gotMessage = message
			gotChoices = append([]string(nil), choices...)
			gotDefaultIndex = defaultIndex
			return "Create a new branch", nil
		},
	}

	mode, err := resolveBranchMode(parsedAddOptions{}, true, ui)
	if err != nil {
		t.Fatalf("resolveBranchMode: %v", err)
	}

	if mode != core.BranchModeNew {
		t.Fatalf("mode = %q, want %q", mode, core.BranchModeNew)
	}
	if gotMessage != "How should the worktree branch be created?" {
		t.Fatalf("message = %q", gotMessage)
	}
	if strings.Join(gotChoices, "|") != "Use an existing branch|Create a new branch" {
		t.Fatalf("choices = %#v", gotChoices)
	}
	if gotDefaultIndex != 0 {
		t.Fatalf("defaultIndex = %d, want 0", gotDefaultIndex)
	}
}

func TestResolveProjectNameAndDirectoryPromptsSequentiallyInInteractiveMode(t *testing.T) {
	t.Parallel()

	var prompts []string
	ui := &promptUI{
		inputPrompter: func(message string, defaultValue string) (string, error) {
			prompts = append(prompts, message+"|"+defaultValue)
			switch len(prompts) {
			case 1:
				return "feature/my-branch2", nil
			case 2:
				return "/tmp/feature-my-branch2", nil
			default:
				t.Fatalf("unexpected prompt count %d", len(prompts))
				return "", nil
			}
		},
	}

	projectName, directory, err := resolveProjectNameAndDirectory(
		"/Users/alex/repos/wtx",
		"feature/my-branch2",
		parsedAddOptions{},
		config.Config{},
		true,
		ui,
	)
	if err != nil {
		t.Fatalf("resolveProjectNameAndDirectory: %v", err)
	}

	if projectName != "feature/my-branch2" {
		t.Fatalf("projectName = %q", projectName)
	}
	if directory != "/tmp/feature-my-branch2" {
		t.Fatalf("directory = %q", directory)
	}
	if len(prompts) != 2 {
		t.Fatalf("prompts = %#v", prompts)
	}
	if prompts[0] != "Project name for the new worktree|feature/my-branch2" {
		t.Fatalf("prompts[0] = %q", prompts[0])
	}
	if prompts[1] != "Directory for the new worktree|/Users/alex/repos/feature-my-branch2" {
		t.Fatalf("prompts[1] = %q", prompts[1])
	}
}

func TestResolveProjectNameAndDirectoryUsesEditedProjectNameForDerivedDirectory(t *testing.T) {
	t.Parallel()

	var prompts []string
	ui := &promptUI{
		inputPrompter: func(message string, defaultValue string) (string, error) {
			prompts = append(prompts, message+"|"+defaultValue)
			switch len(prompts) {
			case 1:
				return "release/1.0", nil
			case 2:
				return "", nil
			default:
				t.Fatalf("unexpected prompt count %d", len(prompts))
				return "", nil
			}
		},
	}

	projectName, directory, err := resolveProjectNameAndDirectory(
		"/Users/alex/repos/wtx",
		"feature/my-branch2",
		parsedAddOptions{},
		config.Config{},
		true,
		ui,
	)
	if err != nil {
		t.Fatalf("resolveProjectNameAndDirectory: %v", err)
	}

	if projectName != "release/1.0" {
		t.Fatalf("projectName = %q", projectName)
	}
	if directory != "/Users/alex/repos/release-1.0" {
		t.Fatalf("directory = %q", directory)
	}
	if prompts[1] != "Directory for the new worktree|/Users/alex/repos/release-1.0" {
		t.Fatalf("prompts[1] = %q", prompts[1])
	}
}

func TestResolveProjectNameAndDirectoryKeepsExplicitDirectoryAfterEditingProjectName(t *testing.T) {
	t.Parallel()

	var prompts []string
	ui := &promptUI{
		inputPrompter: func(message string, defaultValue string) (string, error) {
			prompts = append(prompts, message+"|"+defaultValue)
			switch len(prompts) {
			case 1:
				return "release/1.0", nil
			case 2:
				return "", nil
			default:
				t.Fatalf("unexpected prompt count %d", len(prompts))
				return "", nil
			}
		},
	}

	projectName, directory, err := resolveProjectNameAndDirectory(
		"/Users/alex/repos/wtx",
		"feature/my-branch2",
		parsedAddOptions{directory: "/tmp/custom-worktree"},
		config.Config{},
		true,
		ui,
	)
	if err != nil {
		t.Fatalf("resolveProjectNameAndDirectory: %v", err)
	}

	if projectName != "release/1.0" {
		t.Fatalf("projectName = %q", projectName)
	}
	if directory != "/tmp/custom-worktree" {
		t.Fatalf("directory = %q", directory)
	}
	if prompts[1] != "Directory for the new worktree|/tmp/custom-worktree" {
		t.Fatalf("prompts[1] = %q", prompts[1])
	}
}

func TestResolveProjectNameAndDirectoryNormalizesRelativeInteractiveDirectory(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	ui := &promptUI{
		inputPrompter: func(message string, defaultValue string) (string, error) {
			switch message {
			case "Project name for the new worktree":
				return "feature/my-branch2", nil
			case "Directory for the new worktree":
				return "worktrees/feature-my-branch2", nil
			default:
				t.Fatalf("unexpected prompt %q", message)
				return "", nil
			}
		},
	}

	projectName, directory, err := resolveProjectNameAndDirectory(
		"/Users/alex/repos/wtx",
		"feature/my-branch2",
		parsedAddOptions{},
		config.Config{},
		true,
		ui,
	)
	if err != nil {
		t.Fatalf("resolveProjectNameAndDirectory: %v", err)
	}

	if projectName != "feature/my-branch2" {
		t.Fatalf("projectName = %q", projectName)
	}
	wantDirectory := filepath.Join(cwd, "worktrees/feature-my-branch2")
	if directory != wantDirectory {
		t.Fatalf("directory = %q, want %q", directory, wantDirectory)
	}
}

func TestResolveProjectNameAndDirectoryFallsBackToDefaultProjectNameWhenPromptEmpty(t *testing.T) {
	t.Parallel()

	var prompts []string
	ui := &promptUI{
		inputPrompter: func(message string, defaultValue string) (string, error) {
			prompts = append(prompts, message+"|"+defaultValue)
			return "", nil
		},
	}

	projectName, directory, err := resolveProjectNameAndDirectory(
		"/Users/alex/repos/wtx",
		"feature/my-branch2",
		parsedAddOptions{},
		config.Config{},
		true,
		ui,
	)
	if err != nil {
		t.Fatalf("resolveProjectNameAndDirectory: %v", err)
	}

	if projectName != "feature/my-branch2" {
		t.Fatalf("projectName = %q", projectName)
	}
	if directory != "/Users/alex/repos/feature-my-branch2" {
		t.Fatalf("directory = %q", directory)
	}
	if len(prompts) != 2 {
		t.Fatalf("prompts = %#v", prompts)
	}
	if prompts[0] != "Project name for the new worktree|feature/my-branch2" {
		t.Fatalf("prompts[0] = %q", prompts[0])
	}
	if prompts[1] != "Directory for the new worktree|/Users/alex/repos/feature-my-branch2" {
		t.Fatalf("prompts[1] = %q", prompts[1])
	}
}

func TestResolveProjectNameAndDirectoryDerivesDefaultsWhenNonInteractive(t *testing.T) {
	t.Parallel()

	projectName, directory, err := resolveProjectNameAndDirectory(
		"/Users/alex/repos/wtx",
		"feature/my-branch2",
		parsedAddOptions{},
		config.Config{WorktreeRoot: "/Users/alex/worktrees"},
		false,
		&promptUI{},
	)
	if err != nil {
		t.Fatalf("resolveProjectNameAndDirectory: %v", err)
	}

	if projectName != "feature/my-branch2" {
		t.Fatalf("projectName = %q", projectName)
	}
	if directory != "/Users/alex/worktrees/feature-my-branch2" {
		t.Fatalf("directory = %q", directory)
	}
}
