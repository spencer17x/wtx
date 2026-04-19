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
	if strings.Contains(output, "\x00") {
		return parseWorktreeListPorcelainZ(output)
	}

	return parseWorktreeListPorcelainNewline(output)
}

func parseWorktreeListPorcelainZ(output string) ([]Worktree, error) {
	tokens := strings.Split(output, "\x00")
	worktrees := make([]Worktree, 0, len(tokens)/4+1)
	current := make([]string, 0, 4)

	for _, token := range tokens {
		if token == "" {
			if len(current) == 0 {
				continue
			}

			worktree, err := parseWorktreeRecord(current)
			if err != nil {
				return nil, err
			}

			worktrees = append(worktrees, worktree)
			current = current[:0]
			continue
		}

		current = append(current, token)
	}

	if len(current) > 0 {
		worktree, err := parseWorktreeRecord(current)
		if err != nil {
			return nil, err
		}

		worktrees = append(worktrees, worktree)
	}

	return worktrees, nil
}

func parseWorktreeListPorcelainNewline(output string) ([]Worktree, error) {
	blocks := strings.Split(output, "\n\n")
	worktrees := make([]Worktree, 0, len(blocks))

	for _, block := range blocks {
		block = strings.TrimRight(block, "\r\n")
		if block == "" {
			continue
		}

		worktree, err := parseWorktreeRecord(strings.Split(block, "\n"))
		if err != nil {
			return nil, err
		}

		worktrees = append(worktrees, worktree)
	}

	return worktrees, nil
}

func parseWorktreeRecord(fields []string) (Worktree, error) {
	var worktree Worktree

	for _, field := range fields {
		switch {
		case strings.HasPrefix(field, "worktree "):
			worktree.Path = strings.TrimPrefix(field, "worktree ")
		case strings.HasPrefix(field, "HEAD "):
			worktree.Head = strings.TrimPrefix(field, "HEAD ")
		case strings.HasPrefix(field, "branch "):
			worktree.BranchRef = strings.TrimPrefix(field, "branch ")
		case strings.HasPrefix(field, "prunable"):
			worktree.Prunable = true
		}
	}

	if worktree.Path == "" {
		return Worktree{}, fmt.Errorf("invalid git worktree output: missing worktree path")
	}

	return worktree, nil
}

func ListWorktrees(repoRoot string) ([]Worktree, error) {
	output, err := captureCommand(repoRoot, "git", "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}

	return parseWorktreeListPorcelain(output)
}

func resolveWorktreeByBranch(worktrees []Worktree, branchName string) (Worktree, error) {
	targetRef := "refs/heads/" + branchName
	var prunableMatch bool
	for _, worktree := range worktrees {
		if worktree.BranchRef != targetRef {
			continue
		}
		if worktree.Prunable {
			prunableMatch = true
			continue
		}
		return worktree, nil
	}

	if prunableMatch {
		return Worktree{}, fmt.Errorf("worktree for branch %s is prunable; run git worktree prune", branchName)
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
