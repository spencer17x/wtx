package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spencer17x/wtx/internal/config"
	"github.com/spencer17x/wtx/internal/core"
	"github.com/spencer17x/wtx/internal/fsops"
	wtgit "github.com/spencer17x/wtx/internal/git"
)

type parsedAddOptions struct {
	branchName     string
	batchBranches  []string
	branchMode     core.BranchMode
	baseRef        string
	projectName    string
	directory      string
	worktreeRoot   string
	mode           core.InitializationMode
	yes            bool
	dryRun         bool
	nonInteractive bool
}

type parsedCLI struct {
	command string
	options parsedAddOptions
	target  string
}

func printUsage(stdout io.Writer) {
	fmt.Fprintln(stdout, `Usage:
  wtx add [branch] [options]
  wtx batch-add <branch> [<branch> ...] [options]
  wtx path <branch>
  wtx switch <branch>
  wtx shell-init <shell>

		Options:
		  --new-branch          Create a new branch for the worktree
	  --base <ref>          Base ref when --new-branch is used
	  --name <name>         Worktree project name
	  --dir <path>          Worktree directory
	  --root <path>         Default worktree root directory
	  --mode <mode>         Initialization mode: default | custom | none
	  --dry-run             Print the plan without creating the worktree
	  --non-interactive     Disable prompts even when stdin is a TTY
	  -y, --yes             Accept defaults and skip confirmation prompts
	  -h, --help            Show this help message`)
}

func requireOptionValue(argv []string, index int, option string) (string, error) {
	if index+1 >= len(argv) {
		return "", fmt.Errorf("missing value for %s", option)
	}

	value := argv[index+1]
	if strings.HasPrefix(value, "-") {
		return "", fmt.Errorf("missing value for %s", option)
	}

	return value, nil
}

func parseArgv(argv []string) (parsedCLI, error) {
	if len(argv) == 0 {
		return parsedCLI{command: "help"}, nil
	}

	for _, arg := range argv {
		if arg == "-h" || arg == "--help" {
			return parsedCLI{command: "help"}, nil
		}
	}

	command := argv[0]
	if command != "add" && command != "batch-add" && command != "path" && command != "switch" && command != "shell-init" {
		return parsedCLI{}, fmt.Errorf("unknown command: %s", command)
	}

	if command == "path" || command == "switch" || command == "shell-init" {
		if len(argv) != 2 {
			return parsedCLI{}, fmt.Errorf("%s requires exactly one positional argument", command)
		}
		return parsedCLI{command: command, target: argv[1]}, nil
	}

	options := parsedAddOptions{}
	rest := argv[1:]

	for index := 0; index < len(rest); index++ {
		value := rest[index]

		if !strings.HasPrefix(value, "-") {
			if command == "batch-add" {
				options.batchBranches = append(options.batchBranches, value)
			} else if options.branchName == "" {
				options.branchName = value
			} else {
				return parsedCLI{}, fmt.Errorf("unexpected positional argument: %s", value)
			}
			continue
		}

		switch value {
		case "--new-branch":
			options.branchMode = core.BranchModeNew
		case "--base":
			result, err := requireOptionValue(rest, index, value)
			if err != nil {
				return parsedCLI{}, err
			}
			options.baseRef = result
			index++
		case "--name":
			result, err := requireOptionValue(rest, index, value)
			if err != nil {
				return parsedCLI{}, err
			}
			options.projectName = result
			index++
		case "--dir":
			result, err := requireOptionValue(rest, index, value)
			if err != nil {
				return parsedCLI{}, err
			}
			options.directory = result
			index++
		case "--root":
			result, err := requireOptionValue(rest, index, value)
			if err != nil {
				return parsedCLI{}, err
			}
			options.worktreeRoot = result
			index++
		case "--mode":
			result, err := requireOptionValue(rest, index, value)
			if err != nil {
				return parsedCLI{}, err
			}
			switch core.InitializationMode(result) {
			case core.InitializationModeDefault, core.InitializationModeCustom, core.InitializationModeNone:
				options.mode = core.InitializationMode(result)
			default:
				return parsedCLI{}, fmt.Errorf("unsupported mode: %s", result)
			}
			index++
		case "--dry-run":
			options.dryRun = true
		case "--non-interactive":
			options.nonInteractive = true
		case "-y", "--yes":
			options.yes = true
		default:
			return parsedCLI{}, fmt.Errorf("unknown option: %s", value)
		}
	}

	return parsedCLI{
		command: command,
		options: options,
	}, nil
}

func isInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func shouldUsePrompts(shellInteractive bool, options parsedAddOptions) bool {
	return shellInteractive && !options.yes && !options.nonInteractive
}

func summarizePlan(plan []core.PlanEntry, setupCommands []core.SetupCommand) {
	if len(plan) == 0 {
		fmt.Println("No ignored files or directories were detected for initialization.")
		return
	}

	fmt.Println("Initialization plan:")
	symlinkPaths := []string{}
	for _, entry := range plan {
		fmt.Printf("- %-7s %s\n", entry.Strategy, entry.Path)
		if entry.Strategy == core.StrategySymlink {
			symlinkPaths = append(symlinkPaths, entry.Path)
		}
	}

	if len(setupCommands) > 0 {
		fmt.Println("Setup commands:")
		for _, command := range setupCommands {
			fmt.Printf("- %s %s\n", command.Command, strings.Join(command.Args, " "))
		}
	}

	if len(symlinkPaths) > 0 {
		fmt.Println("Safety note:")
		fmt.Println("- symlink paths stay shared with the source worktree; editing them in either location changes the same underlying files:")
		for _, symlinkPath := range symlinkPaths {
			fmt.Printf("  - %s\n", symlinkPath)
		}
	}
}

func resolveBranchMode(options parsedAddOptions, interactive bool, ui *promptUI) (core.BranchMode, error) {
	if options.branchMode != "" {
		return options.branchMode, nil
	}

	if !interactive || options.yes {
		return core.BranchModeExisting, nil
	}

	choice, err := ui.askSelect(
		"How should the worktree branch be created?",
		[]string{"Use an existing branch", "Create a new branch"},
		1,
	)
	if err != nil {
		return "", err
	}

	if choice == "Create a new branch" {
		return core.BranchModeNew, nil
	}

	return core.BranchModeExisting, nil
}

func resolveBranchName(options parsedAddOptions, branchMode core.BranchMode, currentBranch string, interactive bool, ui *promptUI) (string, error) {
	if options.branchName != "" {
		return options.branchName, nil
	}

	if !interactive || options.yes {
		return "", errors.New("a branch name is required when prompts are disabled")
	}

	message := "Existing branch name"
	if branchMode == core.BranchModeNew {
		message = "New branch name"
	}

	return ui.askInput(message, currentBranch)
}

func resolveBaseRef(options parsedAddOptions, branchMode core.BranchMode, currentBranch string, interactive bool, ui *promptUI) (string, error) {
	if branchMode != core.BranchModeNew {
		return "", nil
	}

	if options.baseRef != "" {
		return options.baseRef, nil
	}

	if !interactive || options.yes {
		if currentBranch == "" {
			return "HEAD", nil
		}
		return currentBranch, nil
	}

	defaultRef := currentBranch
	if defaultRef == "" {
		defaultRef = "HEAD"
	}

	return ui.askInput("Base ref for the new branch", defaultRef)
}

func resolveProjectNameAndDirectory(repoRoot string, branchName string, options parsedAddOptions, cfg config.Config, interactive bool, ui *promptUI) (string, string, error) {
	defaults := core.DeriveWorktreeDefaults(core.WorktreeDefaultsInput{
		BranchName:   branchName,
		RepoRoot:     repoRoot,
		WorktreeRoot: firstNonEmpty(options.worktreeRoot, cfg.WorktreeRoot),
	})

	projectName := firstNonEmpty(options.projectName, defaults.ProjectName)
	directoryInput := options.directory

	if !interactive || options.yes {
		if directoryInput == "" {
			directoryInput = core.BuildDefaultDirectory(projectName, defaults.WorktreeRoot)
		}
	} else {
		var err error
		projectName, err = ui.askInput("Project name for the new worktree", projectName)
		if err != nil {
			return "", "", err
		}
		if projectName == "" {
			projectName = firstNonEmpty(options.projectName, defaults.ProjectName)
		}

		directoryDefault := directoryInput
		if directoryDefault == "" {
			directoryDefault = core.BuildDefaultDirectory(projectName, defaults.WorktreeRoot)
		}

		directoryInput, err = ui.askInput("Directory for the new worktree", directoryDefault)
		if err != nil {
			return "", "", err
		}
		if directoryInput == "" {
			directoryInput = directoryDefault
		}
	}

	directory := directoryInput
	if !filepath.IsAbs(directory) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}

		directory = filepath.Join(cwd, directoryInput)
	}

	return projectName, directory, nil
}

func resolveInitializationPlan(options parsedAddOptions, cfg config.Config, ignoredPaths []string, projectFiles []string, interactive bool, ui *promptUI) ([]core.PlanEntry, []core.SetupCommand, error) {
	if len(ignoredPaths) == 0 {
		return nil, nil, nil
	}

	mode := options.mode
	if mode == "" {
		if !interactive || options.yes {
			mode = core.InitializationModeDefault
		} else {
			choice, err := ui.askSelect(
				"How should ignored files and local environment data be initialized?",
				[]string{"Default plan", "Customize for this run", "Do not reuse anything"},
				0,
			)
			if err != nil {
				return nil, nil, err
			}

			switch choice {
			case "Customize for this run":
				mode = core.InitializationModeCustom
			case "Do not reuse anything":
				mode = core.InitializationModeNone
			default:
				mode = core.InitializationModeDefault
			}
		}
	}

	if mode == core.InitializationModeCustom && (!interactive || options.yes) {
		return nil, nil, errors.New("custom mode requires interactive prompts because each ignored entry needs a strategy")
	}

	customStrategies := map[string]core.Strategy{}
	for key, value := range cfg.StrategyOverrides {
		customStrategies[key] = value
	}
	if mode == core.InitializationModeCustom {
		for _, ignoredPath := range ignoredPaths {
			defaultStrategy := string(core.ClassifyIgnoredPath(ignoredPath))
			if configuredStrategy, ok := cfg.StrategyOverrides[strings.TrimRight(strings.ReplaceAll(ignoredPath, "\\", "/"), "/")]; ok {
				defaultStrategy = string(configuredStrategy)
			}
			choice, err := ui.askSelect(
				fmt.Sprintf("Choose a strategy for %s", ignoredPath),
				[]string{"copy", "symlink", "skip", "setup"},
				map[string]int{"copy": 0, "symlink": 1, "skip": 2, "setup": 3}[defaultStrategy],
			)
			if err != nil {
				return nil, nil, err
			}

			customStrategies[strings.TrimRight(strings.ReplaceAll(ignoredPath, "\\", "/"), "/")] = core.Strategy(choice)
		}
	}

	plan := core.BuildInitializationPlan(core.BuildInitializationPlanInput{
		IgnoredPaths:     ignoredPaths,
		Mode:             mode,
		CustomStrategies: customStrategies,
	})

	setupCommands := []core.SetupCommand{}
	for _, entry := range plan {
		if entry.Strategy == core.StrategySetup {
			setupCommands = core.ApplySetupTemplates(core.DetectSetupCommands(projectFiles), cfg.SetupTemplates)
			break
		}
	}

	return plan, setupCommands, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func hookEnv(repoRoot string, worktreeDirectory string, projectName string, branchName string, branchMode core.BranchMode) map[string]string {
	return map[string]string{
		"WTX_REPO_ROOT":          repoRoot,
		"WTX_WORKTREE_DIRECTORY": worktreeDirectory,
		"WTX_PROJECT_NAME":       projectName,
		"WTX_BRANCH_NAME":        branchName,
		"WTX_BRANCH_MODE":        string(branchMode),
	}
}

func runCommandWithEnv(cwd string, env map[string]string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed (%s %s): %w", command, strings.Join(args, " "), err)
	}

	return nil
}

func runCommand(cwd string, command string, args ...string) error {
	return runCommandWithEnv(cwd, nil, command, args...)
}

func executePlan(
	stdout io.Writer,
	options parsedAddOptions,
	repoRoot string,
	projectName string,
	branchName string,
	branchMode core.BranchMode,
	gitCommand core.GitCommand,
	applyInput fsops.ApplyInitializationPlanInput,
	hooks map[string][]core.CommandHook,
	runGit func(string, string, ...string) error,
	applyInitialization func(fsops.ApplyInitializationPlanInput) error,
	runHook func(string, map[string]string, string, ...string) error,
) error {
	env := hookEnv(repoRoot, applyInput.WorktreeDirectory, projectName, branchName, branchMode)
	if options.dryRun {
		fmt.Fprintln(stdout, "Dry run:")
		for _, hook := range hooks["beforeCreate"] {
			fmt.Fprintf(stdout, "- would run hook (%s): %s %s\n", hook.Description, hook.Command, strings.Join(hook.Args, " "))
		}
		fmt.Fprintf(stdout, "- would run: %s %s\n", gitCommand.Command, strings.Join(gitCommand.Args, " "))
		if len(applyInput.Plan) > 0 || len(applyInput.SetupCommands) > 0 {
			fmt.Fprintf(stdout, "- would initialize the new worktree at %s\n", applyInput.WorktreeDirectory)
		}
		for _, hook := range hooks["afterCreate"] {
			fmt.Fprintf(stdout, "- would run hook (%s): %s %s\n", hook.Description, hook.Command, strings.Join(hook.Args, " "))
		}
		fmt.Fprintln(stdout, "No changes were made.")
		return nil
	}

	for _, hook := range hooks["beforeCreate"] {
		fmt.Fprintf(stdout, "Running hook: %s\n", hook.Description)
		if err := runHook(repoRoot, env, hook.Command, hook.Args...); err != nil {
			return err
		}
	}

	fmt.Fprintf(stdout, "Running: %s %s\n", gitCommand.Command, strings.Join(gitCommand.Args, " "))
	if err := runGit(repoRoot, gitCommand.Command, gitCommand.Args...); err != nil {
		return err
	}

	if err := applyInitialization(applyInput); err != nil {
		return err
	}

	for _, hook := range hooks["afterCreate"] {
		fmt.Fprintf(stdout, "Running hook: %s\n", hook.Description)
		if err := runHook(applyInput.WorktreeDirectory, env, hook.Command, hook.Args...); err != nil {
			return err
		}
	}

	fmt.Fprintln(stdout, "Worktree created and initialized successfully.")
	return nil
}

func runSingleAdd(
	stdout io.Writer,
	options parsedAddOptions,
	cfg config.Config,
	repoRoot string,
	currentBranch string,
	branchName string,
	interactive bool,
	ui *promptUI,
	ignoredPaths []string,
	projectFiles []string,
) error {
	baseRef, err := resolveBaseRef(options, options.branchMode, currentBranch, interactive, ui)
	if err != nil {
		return err
	}

	projectName, directory, err := resolveProjectNameAndDirectory(repoRoot, branchName, options, cfg, interactive, ui)
	if err != nil {
		return err
	}

	plan, setupCommands, err := resolveInitializationPlan(options, cfg, ignoredPaths, projectFiles, interactive, ui)
	if err != nil {
		return err
	}

	gitCommand := wtgit.BuildWorktreeAddCommand(wtgit.BuildWorktreeAddCommandInput{
		BranchMode: options.branchMode,
		BranchName: branchName,
		Directory:  directory,
		BaseRef:    baseRef,
	})

	fmt.Fprintf(stdout, "Project name: %s\n", projectName)
	fmt.Fprintf(stdout, "Worktree directory: %s\n", directory)
	fmt.Fprintf(stdout, "Branch mode: %s\n", options.branchMode)
	summarizePlan(plan, setupCommands)

	if interactive && !options.dryRun {
		confirmed, confirmErr := ui.askConfirm("Create the worktree with this plan?", true)
		if confirmErr != nil {
			return confirmErr
		}

		if !confirmed {
			fmt.Fprintln(stdout, "Cancelled.")
			return nil
		}
	}

	return executePlan(stdout, options, repoRoot, projectName, branchName, options.branchMode, gitCommand, fsops.ApplyInitializationPlanInput{
		SourceRoot:        repoRoot,
		WorktreeDirectory: directory,
		Plan:              plan,
		SetupCommands:     setupCommands,
	}, cfg.Hooks, runCommand, fsops.ApplyInitializationPlan, runCommandWithEnv)
}

func runBatchAdd(
	stdout io.Writer,
	options parsedAddOptions,
	cfg config.Config,
	repoRoot string,
	currentBranch string,
	ignoredPaths []string,
	projectFiles []string,
) error {
	if len(options.batchBranches) == 0 {
		return errors.New("batch-add requires at least one branch")
	}
	if options.projectName != "" || options.directory != "" {
		return errors.New("batch-add does not support --name or --dir; use --root to control the parent directory")
	}

	batchOptions := options
	if batchOptions.branchMode == "" {
		batchOptions.branchMode = core.BranchModeExisting
	}

	for _, branchName := range options.batchBranches {
		fmt.Fprintf(stdout, "==> %s\n", branchName)
		currentOptions := batchOptions
		currentOptions.branchName = branchName
		if err := runSingleAdd(stdout, currentOptions, cfg, repoRoot, currentBranch, branchName, false, newPromptUI(), ignoredPaths, projectFiles); err != nil {
			return err
		}
	}

	return nil
}

func Run(argv []string) error {
	parsed, err := parseArgv(argv)
	if err != nil {
		return err
	}

	if parsed.command == "help" {
		printUsage(os.Stdout)
		return nil
	}

	options := parsed.options
	interactive := shouldUsePrompts(isInteractive(), options)
	ui := newPromptUI()

	if parsed.command == "shell-init" {
		return runShellInit(os.Stdout, parsed.target)
	}

	repoRoot, err := wtgit.ResolveRepoRoot(".")
	if err != nil {
		return err
	}

	resolvePath := func(branchName string) (string, error) {
		worktree, err := wtgit.FindWorktreeByBranch(repoRoot, branchName)
		if err != nil {
			return "", err
		}
		return worktree.Path, nil
	}

	if parsed.command == "path" {
		return runPath(os.Stdout, parsed.target, resolvePath)
	}
	if parsed.command == "switch" {
		return runSwitch(os.Stdout, parsed.target, resolvePath)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cfg, err := config.LoadMergedConfig(repoRoot, homeDir)
	if err != nil {
		return err
	}

	currentBranch, err := wtgit.ResolveCurrentBranch(repoRoot)
	if err != nil {
		return err
	}

	ignoredPaths, err := wtgit.ListIgnoredPaths(repoRoot)
	if err != nil {
		return err
	}

	projectFiles, err := wtgit.ListProjectFiles(repoRoot)
	if err != nil {
		return err
	}

	if parsed.command == "batch-add" {
		return runBatchAdd(os.Stdout, options, cfg, repoRoot, currentBranch, ignoredPaths, projectFiles)
	}

	branchMode, err := resolveBranchMode(options, interactive, ui)
	if err != nil {
		return err
	}

	branchName, err := resolveBranchName(options, branchMode, currentBranch, interactive, ui)
	if err != nil {
		return err
	}

	baseRef, err := resolveBaseRef(options, branchMode, currentBranch, interactive, ui)
	if err != nil {
		return err
	}
	options.branchMode = branchMode
	options.baseRef = baseRef

	return runSingleAdd(os.Stdout, options, cfg, repoRoot, currentBranch, branchName, interactive, ui, ignoredPaths, projectFiles)
}
