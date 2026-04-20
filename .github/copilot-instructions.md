# Copilot Instructions

## Build & Test

```bash
# Build the binary
go build -o bin/wtx ./cmd/wtx

# Run all tests (Go + npm)
npm test

# Run only Go tests
go test ./...

# Run a single Go test
go test ./internal/cli/... -run TestParseArgvRecognizesDryRunAndNonInteractive

# Run only npm tests
npm run test:npm

# Validate GoReleaser config
npm run release:check
```

## Architecture

`wtx` is a Go CLI that wraps `git worktree add` with local environment initialization. The flow is:

1. **`cmd/wtx/main.go`** — entrypoint, calls `cli.Run(os.Args[1:])`
2. **`internal/cli/`** — orchestrates everything: argument parsing (`parseArgv`), interactive prompts, and execution. `app.go` is the main file; `Run()` is the top-level function.
3. **`internal/core/`** — pure business logic with no I/O:
   - `types.go` — all shared types (`Strategy`, `BranchMode`, `InitializationMode`, `PlanEntry`, `SetupCommand`, `CommandHook`, `GitCommand`)
   - `planner.go` — `BuildInitializationPlan` and `ClassifyIgnoredPath` (assigns copy/symlink/skip/setup per ignored path)
   - `setup.go` — `DetectSetupCommands` (ecosystem detection) and `ApplySetupTemplates` (config overrides)
   - `defaults.go` — `DeriveWorktreeDefaults` and `BuildDefaultDirectory`
4. **`internal/config/`** — loads and merges `~/.wtx.json` and `<repo>/.wtx.json`; raw JSON types (`rawConfig`, `rawSetupCommand`) are parsed into typed `Config`/`SetupCommand`/`CommandHook` structs
5. **`internal/git/`** — git command construction (`BuildWorktreeAddCommand`) and repo inspection (`ResolveRepoRoot`, `ListIgnoredPaths`, `ListProjectFiles`, `FindWorktreeByBranch`)
6. **`internal/fsops/apply.go`** — executes the plan: copy, symlink, and setup commands in the new worktree

Config merges user-level (`~/.wtx.json`) first, then project-level (`<repo>/.wtx.json`) on top.

## Key Conventions

**Testing style:** Standard library `testing` only — no third-party test frameworks. All test functions call `t.Parallel()` and check errors manually with `t.Fatalf`/`t.Errorf` (no assertion helpers).

**Dependency injection in `executePlan`:** The functions `runGit`, `applyInitialization`, and `runHook` are passed as parameters rather than called directly, enabling unit testing without real filesystem or Git side effects.

**Path normalization:** Every path key that comes from user input or git output is normalized via `normalizeCandidatePath` (backslashes → forward slashes, strip leading `./`, strip trailing `/`). Always normalize before map lookups.

**Type definitions live in `internal/core/types.go`:** All packages that need shared types import `core`. Avoid duplicating type definitions elsewhere.

**`firstNonEmpty` helper:** Used throughout CLI resolution functions to express precedence (flag > config > derived default) concisely.

**Config validation pattern:** Raw JSON structs (`rawConfig`, `rawSetupCommand`) are parsed into typed structs with explicit validation; any invalid field (empty `id`, `description`, `command`, or unknown strategy string) returns a descriptive error referencing the config file path.

**npm wrapper for distribution:** The binary is distributed via npm (`@spencer17x/wtx`). The npm package in `npm/` downloads and installs the platform-specific Go binary. The CI pipeline runs `npm test` (which includes `go test ./...`) rather than running Go directly.
