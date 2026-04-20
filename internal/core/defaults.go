package core

import (
	"path/filepath"
	"regexp"
	"strings"
)

type WorktreeDefaultsInput struct {
	BranchName   string
	RepoRoot     string
	WorktreeRoot string
}

type WorktreeDefaults struct {
	ProjectName  string
	WorktreeRoot string
	Directory    string
}

type ProjectLocation struct {
	ProjectName  string
	WorktreeRoot string
	Directory    string
}

type ProjectLocationEdit struct {
	ProjectName   string
	Directory     string
	EditProject   bool
	EditDirectory bool
}

var dashPattern = regexp.MustCompile(`-+`)

func SanitizeDirectoryName(value string) string {
	sanitized := strings.TrimSpace(value)
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	sanitized = strings.Join(strings.Fields(sanitized), "-")
	sanitized = dashPattern.ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	if sanitized == "" {
		return "worktree"
	}

	return sanitized
}

func BuildDefaultDirectory(projectName string, worktreeRoot string) string {
	return filepath.Join(worktreeRoot, SanitizeDirectoryName(projectName))
}

func DeriveWorktreeDefaults(input WorktreeDefaultsInput) WorktreeDefaults {
	worktreeRoot := input.WorktreeRoot
	if worktreeRoot == "" {
		worktreeRoot = filepath.Dir(input.RepoRoot)
	}

	return WorktreeDefaults{
		ProjectName:  input.BranchName,
		WorktreeRoot: worktreeRoot,
		Directory:    BuildDefaultDirectory(input.BranchName, worktreeRoot),
	}
}

func ApplyProjectLocationEdit(current ProjectLocation, edit ProjectLocationEdit) ProjectLocation {
	next := current

	if edit.EditProject {
		next.ProjectName = edit.ProjectName
		next.Directory = BuildDefaultDirectory(next.ProjectName, next.WorktreeRoot)
	}

	if edit.EditDirectory {
		next.Directory = edit.Directory
	}

	return next
}
