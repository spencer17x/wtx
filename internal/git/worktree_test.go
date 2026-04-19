package git

import "testing"

func TestParseWorktreeListPorcelainIncludesBranchAndPrunable(t *testing.T) {
	t.Parallel()

	output := `worktree /repo
HEAD 1111111111111111111111111111111111111111
branch refs/heads/main

worktree /repo-feature
HEAD 2222222222222222222222222222222222222222
branch refs/heads/feature/demo
prunable gitdir file points to non-existent location
`

	worktrees, err := parseWorktreeListPorcelain(output)
	if err != nil {
		t.Fatalf("parseWorktreeListPorcelain: %v", err)
	}

	if len(worktrees) != 2 {
		t.Fatalf("len(worktrees) = %d, want 2", len(worktrees))
	}
	if worktrees[0].Path != "/repo" || worktrees[0].BranchRef != "refs/heads/main" {
		t.Fatalf("worktrees[0] = %#v", worktrees[0])
	}
	if !worktrees[1].Prunable {
		t.Fatalf("worktrees[1] = %#v, want prunable", worktrees[1])
	}
}

func TestResolveWorktreeByBranchRequiresExactNonPrunableMatch(t *testing.T) {
	t.Parallel()

	worktrees := []Worktree{
		{Path: "/repo", BranchRef: "refs/heads/main"},
		{Path: "/repo-feature", BranchRef: "refs/heads/feature/demo"},
		{Path: "/repo-feature-old", BranchRef: "refs/heads/feature/demo-old", Prunable: true},
		{Path: "/repo-detached"},
	}

	worktree, err := resolveWorktreeByBranch(worktrees, "feature/demo")
	if err != nil {
		t.Fatalf("resolveWorktreeByBranch: %v", err)
	}
	if worktree.Path != "/repo-feature" {
		t.Fatalf("worktree.Path = %q, want %q", worktree.Path, "/repo-feature")
	}

	if _, err := resolveWorktreeByBranch(worktrees, "feature"); err == nil {
		t.Fatal("expected exact branch lookup to reject fuzzy matches")
	}

	if _, err := resolveWorktreeByBranch([]Worktree{
		{Path: "/repo-feature-old", BranchRef: "refs/heads/feature/demo", Prunable: true},
	}, "feature/demo"); err == nil {
		t.Fatal("expected prunable worktree lookup to fail")
	}
}

func TestResolveWorktreeByBranchPrefersHealthyExactMatchOverPrunable(t *testing.T) {
	t.Parallel()

	worktrees := []Worktree{
		{Path: "/repo-feature-prunable", BranchRef: "refs/heads/feature/demo", Prunable: true},
		{Path: "/repo-feature", BranchRef: "refs/heads/feature/demo"},
	}

	worktree, err := resolveWorktreeByBranch(worktrees, "feature/demo")
	if err != nil {
		t.Fatalf("resolveWorktreeByBranch: %v", err)
	}

	if worktree.Path != "/repo-feature" {
		t.Fatalf("worktree.Path = %q, want %q", worktree.Path, "/repo-feature")
	}
}

func TestParseWorktreeListPorcelainZHandlesNewlineInPath(t *testing.T) {
	t.Parallel()

	output := "worktree /repo\x00HEAD 1111111111111111111111111111111111111111\x00branch refs/heads/main\x00\x00worktree /repo\nfeature\x00HEAD 2222222222222222222222222222222222222222\x00branch refs/heads/feature/demo\x00"

	worktrees, err := parseWorktreeListPorcelain(output)
	if err != nil {
		t.Fatalf("parseWorktreeListPorcelain: %v", err)
	}

	if len(worktrees) != 2 {
		t.Fatalf("len(worktrees) = %d, want 2", len(worktrees))
	}
	if worktrees[1].Path != "/repo\nfeature" {
		t.Fatalf("worktrees[1].Path = %q, want %q", worktrees[1].Path, "/repo\nfeature")
	}
	if worktrees[1].BranchRef != "refs/heads/feature/demo" {
		t.Fatalf("worktrees[1].BranchRef = %q, want %q", worktrees[1].BranchRef, "refs/heads/feature/demo")
	}
}
