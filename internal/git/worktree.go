package git

import (
	"fmt"
	"strings"
)

type Worktree struct {
	Path      string
	Head      string
	BranchRef string
	Prunable  bool
}

func parseWorktreeListPorcelain(output string) ([]Worktree, error) {
	blocks := strings.Split(strings.TrimSpace(output), "\n\n")
	worktrees := make([]Worktree, 0, len(blocks))

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		var worktree Worktree
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "worktree "):
				worktree.Path = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			case strings.HasPrefix(line, "HEAD "):
				worktree.Head = strings.TrimSpace(strings.TrimPrefix(line, "HEAD "))
			case strings.HasPrefix(line, "branch "):
				worktree.BranchRef = strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			case strings.HasPrefix(line, "prunable"):
				worktree.Prunable = true
			}
		}

		if worktree.Path == "" {
			return nil, fmt.Errorf("invalid git worktree output: missing worktree path")
		}

		worktrees = append(worktrees, worktree)
	}

	return worktrees, nil
}

func ListWorktrees(repoRoot string) ([]Worktree, error) {
	output, err := captureCommand(repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseWorktreeListPorcelain(output)
}

func resolveWorktreeByBranch(worktrees []Worktree, branchName string) (Worktree, error) {
	targetRef := "refs/heads/" + branchName
	for _, worktree := range worktrees {
		if worktree.BranchRef != targetRef {
			continue
		}
		if worktree.Prunable {
			return Worktree{}, fmt.Errorf("worktree for branch %s is prunable; run git worktree prune", branchName)
		}
		return worktree, nil
	}

	return Worktree{}, fmt.Errorf("no worktree found for branch: %s", branchName)
}

func FindWorktreeByBranch(repoRoot string, branchName string) (Worktree, error) {
	worktrees, err := ListWorktrees(repoRoot)
	if err != nil {
		return Worktree{}, err
	}

	return resolveWorktreeByBranch(worktrees, branchName)
}
