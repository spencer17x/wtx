package core

type BranchMode string

const (
	BranchModeExisting BranchMode = "existing"
	BranchModeNew      BranchMode = "new"
)

type InitializationMode string

const (
	InitializationModeDefault InitializationMode = "default"
	InitializationModeCustom  InitializationMode = "custom"
	InitializationModeNone    InitializationMode = "none"
)

type Strategy string

const (
	StrategyCopy    Strategy = "copy"
	StrategySymlink Strategy = "symlink"
	StrategySkip    Strategy = "skip"
	StrategySetup   Strategy = "setup"
)

type GitCommand struct {
	Command string
	Args    []string
}

type PlanEntry struct {
	Path     string
	Strategy Strategy
}

type SetupCommand struct {
	ID          string
	Description string
	Command     string
	Args        []string
}

type CommandHook struct {
	ID          string
	Description string
	Command     string
	Args        []string
}
