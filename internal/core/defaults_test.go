package core_test

import (
	"testing"

	"github.com/spencer17x/wtx/internal/core"
)

func TestDeriveWorktreeDefaultsUsesBranchNameAndSanitizedDirectory(t *testing.T) {
	t.Parallel()

	got := core.DeriveWorktreeDefaults(core.WorktreeDefaultsInput{
		BranchName: "feature/awesome-ui",
		RepoRoot:   "/Users/alex/repos/wtx",
	})

	if got.ProjectName != "feature/awesome-ui" {
		t.Fatalf("project name = %q, want %q", got.ProjectName, "feature/awesome-ui")
	}

	if got.WorktreeRoot != "/Users/alex/repos" {
		t.Fatalf("worktree root = %q, want %q", got.WorktreeRoot, "/Users/alex/repos")
	}

	if got.Directory != "/Users/alex/repos/feature-awesome-ui" {
		t.Fatalf("directory = %q, want %q", got.Directory, "/Users/alex/repos/feature-awesome-ui")
	}
}

func TestDeriveWorktreeDefaultsUsesCustomWorktreeRoot(t *testing.T) {
	t.Parallel()

	got := core.DeriveWorktreeDefaults(core.WorktreeDefaultsInput{
		BranchName:   "release-1.0",
		RepoRoot:     "/Users/alex/repos/wtx",
		WorktreeRoot: "/Users/alex/worktrees",
	})

	if got.ProjectName != "release-1.0" {
		t.Fatalf("project name = %q, want %q", got.ProjectName, "release-1.0")
	}

	if got.WorktreeRoot != "/Users/alex/worktrees" {
		t.Fatalf("worktree root = %q, want %q", got.WorktreeRoot, "/Users/alex/worktrees")
	}

	if got.Directory != "/Users/alex/worktrees/release-1.0" {
		t.Fatalf("directory = %q, want %q", got.Directory, "/Users/alex/worktrees/release-1.0")
	}
}

func TestApplyProjectLocationEditRebuildsDirectoryWhenProjectNameChanges(t *testing.T) {
	t.Parallel()

	got := core.ApplyProjectLocationEdit(core.ProjectLocation{
		ProjectName:  "feature/awesome-ui",
		WorktreeRoot: "/Users/alex/repos",
		Directory:    "/Users/alex/repos/feature-awesome-ui",
	}, core.ProjectLocationEdit{
		ProjectName: "release/1.0",
		EditProject: true,
	})

	if got.ProjectName != "release/1.0" {
		t.Fatalf("project name = %q, want %q", got.ProjectName, "release/1.0")
	}

	if got.Directory != "/Users/alex/repos/release-1.0" {
		t.Fatalf("directory = %q, want %q", got.Directory, "/Users/alex/repos/release-1.0")
	}
}

func TestApplyProjectLocationEditAllowsExplicitDirectoryOverride(t *testing.T) {
	t.Parallel()

	got := core.ApplyProjectLocationEdit(core.ProjectLocation{
		ProjectName:  "feature/awesome-ui",
		WorktreeRoot: "/Users/alex/repos",
		Directory:    "/Users/alex/repos/feature-awesome-ui",
	}, core.ProjectLocationEdit{
		ProjectName:   "release/1.0",
		Directory:     "/tmp/custom-dir",
		EditProject:   true,
		EditDirectory: true,
	})

	if got.ProjectName != "release/1.0" {
		t.Fatalf("project name = %q, want %q", got.ProjectName, "release/1.0")
	}

	if got.Directory != "/tmp/custom-dir" {
		t.Fatalf("directory = %q, want %q", got.Directory, "/tmp/custom-dir")
	}
}
