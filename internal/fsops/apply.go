package fsops

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spencer17x/wtx/internal/core"
)

type ApplyInitializationPlanInput struct {
	SourceRoot        string
	WorktreeDirectory string
	Plan              []core.PlanEntry
	SetupCommands     []core.SetupCommand
}

func pathExists(targetPath string) bool {
	_, err := os.Lstat(targetPath)
	return err == nil
}

func ensureParent(targetPath string) error {
	return os.MkdirAll(filepath.Dir(targetPath), 0o755)
}

func copyFile(sourcePath string, targetPath string, info fs.FileInfo) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := ensureParent(targetPath); err != nil {
		return err
	}

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	return err
}

func copyEntry(sourcePath string, targetPath string) error {
	if !pathExists(sourcePath) || pathExists(targetPath) {
		return nil
	}

	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(sourcePath, func(current string, currentInfo fs.FileInfo, currentErr error) error {
			if currentErr != nil {
				return currentErr
			}

			relativePath, err := filepath.Rel(sourcePath, current)
			if err != nil {
				return err
			}

			targetCurrent := targetPath
			if relativePath != "." {
				targetCurrent = filepath.Join(targetPath, relativePath)
			}

			if currentInfo.IsDir() {
				return os.MkdirAll(targetCurrent, currentInfo.Mode())
			}

			if currentInfo.Mode()&os.ModeSymlink != 0 {
				linkTarget, err := os.Readlink(current)
				if err != nil {
					return err
				}

				if err := ensureParent(targetCurrent); err != nil {
					return err
				}

				return os.Symlink(linkTarget, targetCurrent)
			}

			return copyFile(current, targetCurrent, currentInfo)
		})
	}

	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(sourcePath)
		if err != nil {
			return err
		}

		if err := ensureParent(targetPath); err != nil {
			return err
		}

		return os.Symlink(linkTarget, targetPath)
	}

	return copyFile(sourcePath, targetPath, info)
}

func symlinkEntry(sourcePath string, targetPath string) error {
	if !pathExists(sourcePath) || pathExists(targetPath) {
		return nil
	}

	if err := ensureParent(targetPath); err != nil {
		return err
	}

	return os.Symlink(sourcePath, targetPath)
}

func runCommand(cwd string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed (%s %v): %w", command, args, err)
	}

	return nil
}

func ApplyInitializationPlan(input ApplyInitializationPlanInput) error {
	for _, entry := range input.Plan {
		sourcePath := filepath.Join(input.SourceRoot, filepath.FromSlash(entry.Path))
		targetPath := filepath.Join(input.WorktreeDirectory, filepath.FromSlash(entry.Path))

		switch entry.Strategy {
		case core.StrategyCopy:
			if err := copyEntry(sourcePath, targetPath); err != nil {
				return err
			}
		case core.StrategySymlink:
			if err := symlinkEntry(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	for _, setupCommand := range input.SetupCommands {
		fmt.Printf("Running setup: %s\n", setupCommand.Description)
		if err := runCommand(input.WorktreeDirectory, setupCommand.Command, setupCommand.Args...); err != nil {
			return err
		}
	}

	return nil
}
