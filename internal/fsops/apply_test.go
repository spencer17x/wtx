package fsops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spencer17x/wtx/internal/core"
)

func TestApplyInitializationPlanCopiesFilesDirectoriesAndSymlinks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	worktreeDirectory := filepath.Join(root, "worktree")

	if err := os.MkdirAll(filepath.Join(sourceRoot, ".vscode"), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(worktreeDirectory, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, ".env"), []byte("TOKEN=source\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, ".vscode", "settings.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}
	if err := os.Symlink("../shared.json", filepath.Join(sourceRoot, ".vscode", "shared.json")); err != nil {
		t.Fatalf("symlink settings: %v", err)
	}

	err := ApplyInitializationPlan(ApplyInitializationPlanInput{
		SourceRoot:        sourceRoot,
		WorktreeDirectory: worktreeDirectory,
		Plan: []core.PlanEntry{
			{Path: ".env", Strategy: core.StrategyCopy},
			{Path: ".vscode", Strategy: core.StrategyCopy},
		},
	})
	if err != nil {
		t.Fatalf("apply initialization plan: %v", err)
	}

	envData, err := os.ReadFile(filepath.Join(worktreeDirectory, ".env"))
	if err != nil {
		t.Fatalf("read copied env: %v", err)
	}
	if string(envData) != "TOKEN=source\n" {
		t.Fatalf("copied env = %q", envData)
	}

	settingsData, err := os.ReadFile(filepath.Join(worktreeDirectory, ".vscode", "settings.json"))
	if err != nil {
		t.Fatalf("read copied settings: %v", err)
	}
	if string(settingsData) != "{}\n" {
		t.Fatalf("copied settings = %q", settingsData)
	}

	linkTarget, err := os.Readlink(filepath.Join(worktreeDirectory, ".vscode", "shared.json"))
	if err != nil {
		t.Fatalf("read copied symlink: %v", err)
	}
	if linkTarget != "../shared.json" {
		t.Fatalf("copied symlink target = %q", linkTarget)
	}
}

func TestApplyInitializationPlanSkipsExistingTargets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	worktreeDirectory := filepath.Join(root, "worktree")

	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(worktreeDirectory, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, ".env"), []byte("TOKEN=source\n"), 0o600); err != nil {
		t.Fatalf("write source env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreeDirectory, ".env"), []byte("TOKEN=target\n"), 0o600); err != nil {
		t.Fatalf("write target env: %v", err)
	}

	err := ApplyInitializationPlan(ApplyInitializationPlanInput{
		SourceRoot:        sourceRoot,
		WorktreeDirectory: worktreeDirectory,
		Plan: []core.PlanEntry{
			{Path: ".env", Strategy: core.StrategyCopy},
		},
	})
	if err != nil {
		t.Fatalf("apply initialization plan: %v", err)
	}

	envData, err := os.ReadFile(filepath.Join(worktreeDirectory, ".env"))
	if err != nil {
		t.Fatalf("read target env: %v", err)
	}
	if string(envData) != "TOKEN=target\n" {
		t.Fatalf("target env = %q, want existing file to remain", envData)
	}
}

func TestApplyInitializationPlanRunsSetupCommandsInWorkingDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	worktreeDirectory := filepath.Join(root, "worktree")
	frontendDirectory := filepath.Join(worktreeDirectory, "frontend")

	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(frontendDirectory, 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}

	err := ApplyInitializationPlan(ApplyInitializationPlanInput{
		SourceRoot:        sourceRoot,
		WorktreeDirectory: worktreeDirectory,
		SetupCommands: []core.SetupCommand{
			{
				ID:               "record-pwd",
				Description:      "Record setup cwd",
				Command:          "sh",
				Args:             []string{"-c", "pwd > setup-pwd.txt"},
				WorkingDirectory: "frontend",
			},
		},
	})
	if err != nil {
		t.Fatalf("apply initialization plan: %v", err)
	}

	pwdData, err := os.ReadFile(filepath.Join(frontendDirectory, "setup-pwd.txt"))
	if err != nil {
		t.Fatalf("read setup pwd: %v", err)
	}
	if strings.TrimSpace(string(pwdData)) != frontendDirectory {
		t.Fatalf("setup cwd = %q, want %q", strings.TrimSpace(string(pwdData)), frontendDirectory)
	}
}
