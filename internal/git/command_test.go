package git_test

import (
	"reflect"
	"testing"

	"github.com/spencer17x/wtx/internal/core"
	"github.com/spencer17x/wtx/internal/git"
)

func TestBuildWorktreeAddCommandForExistingBranch(t *testing.T) {
	t.Parallel()

	got := git.BuildWorktreeAddCommand(git.BuildWorktreeAddCommandInput{
		BranchMode: core.BranchModeExisting,
		BranchName: "feature/cli",
		Directory:  "/tmp/feature-cli",
	})

	want := core.GitCommand{
		Command: "git",
		Args:    []string{"worktree", "add", "/tmp/feature-cli", "feature/cli"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func TestBuildWorktreeAddCommandForNewBranch(t *testing.T) {
	t.Parallel()

	got := git.BuildWorktreeAddCommand(git.BuildWorktreeAddCommandInput{
		BranchMode: core.BranchModeNew,
		BranchName: "feature/cli",
		BaseRef:    "origin/main",
		Directory:  "/tmp/feature-cli",
	})

	want := core.GitCommand{
		Command: "git",
		Args:    []string{"worktree", "add", "-b", "feature/cli", "/tmp/feature-cli", "origin/main"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}
