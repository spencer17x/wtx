package git

import "github.com/spencer17x/wtx/internal/core"

type BuildWorktreeAddCommandInput struct {
	BranchMode core.BranchMode
	BranchName string
	Directory  string
	BaseRef    string
}

func BuildWorktreeAddCommand(input BuildWorktreeAddCommandInput) core.GitCommand {
	if input.BranchMode == core.BranchModeNew {
		baseRef := input.BaseRef
		if baseRef == "" {
			baseRef = "HEAD"
		}

		return core.GitCommand{
			Command: "git",
			Args: []string{
				"worktree",
				"add",
				"-b",
				input.BranchName,
				input.Directory,
				baseRef,
			},
		}
	}

	return core.GitCommand{
		Command: "git",
		Args:    []string{"worktree", "add", input.Directory, input.BranchName},
	}
}
