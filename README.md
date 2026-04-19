# wtx

[中文说明](./README.zh-CN.md)

`wtx` is a Go CLI that extends `git worktree add` with local environment setup.

Creating a new worktree is usually only the first step. Teams still need to copy local config, reuse editor folders, rebuild dependency environments, and avoid carrying over build artifacts. `wtx` sits on top of Git’s native worktree flow and handles that initialization layer.

Release guide: [docs/releasing.md](./docs/releasing.md)

## What It Does

- Wraps `git worktree add` instead of replacing it
- Creates worktrees from existing branches or new branches
- Derives sensible defaults for project name and worktree directory
- Reuses ignored files and directories with explicit strategies:
  - `copy`
  - `symlink`
  - `skip`
  - `setup`
- Detects common project types and runs matching setup commands
- Supports one-off overrides, persistent config, hooks, dry runs, and batch creation

## Install

### npm

Requirements: supports macOS/Linux on `x64` and `arm64`, and requires Node.js 18+ with npm.

```bash
npm install -g @spencer17x/wtx
```

The npm package name is scoped as `@spencer17x/wtx`, but the installed command remains `wtx`.

### Build from source

Requirements: Go must be installed locally.

```bash
go build -o bin/wtx ./cmd/wtx
```

This produces `./bin/wtx` if you want to run the locally built binary directly.

Run tests:

```bash
go test ./...
```

## Quick Start

Create a worktree from an existing branch:

```bash
wtx add feature/my-branch
```

Create a worktree and a new branch:

```bash
wtx add feature/my-branch --new-branch --base main
```

Preview the plan without changing anything:

```bash
wtx add feature/my-branch --new-branch --dry-run --non-interactive
```

Create multiple worktrees in one command:

```bash
wtx batch-add feature/one feature/two --new-branch --root ~/worktrees --non-interactive
```

Find the worktree path for a branch:

```bash
wtx path feature/my-branch
```

Resolve a branch worktree and print the guidance to switch into it:

```bash
wtx switch feature/my-branch
```

Enable optional shell integration in zsh so `wtx switch` can change your current shell directly:

```bash
eval "$(wtx shell-init zsh)"
```

## Commands

### `add`

```bash
wtx add <branch> [options]
```

Creates a single worktree.

### `batch-add`

```bash
wtx batch-add <branch> [<branch> ...] [options]
```

Creates multiple worktrees in sequence using the same options.

Notes:

- `batch-add` is intended for non-interactive use
- `batch-add` does not support `--name` or `--dir`
- Use `--root` to control the parent worktree directory for batch creation

### `path`

```bash
wtx path <branch>
```

Prints the absolute worktree path for the given branch.

### `switch`

```bash
wtx switch <branch>
```

Without shell integration, prints the target worktree path plus guidance to run `cd "$(wtx path <branch>)"`. With shell integration enabled, it can change your current shell directly.

### `shell-init`

```bash
wtx shell-init <shell>
```

Prints shell integration code for supported shells.

Example for zsh:

```bash
eval "$(wtx shell-init zsh)"
```

## Options

- `--new-branch`
  Create a new branch for the worktree
- `--base <ref>`
  Base ref to use with `--new-branch`
- `--name <name>`
  Override the derived project name
- `--dir <path>`
  Override the target worktree directory
- `--root <path>`
  Override the default worktree root directory
- `--mode <default|custom|none>`
  Choose the initialization mode
- `--dry-run`
  Print the full plan without creating the worktree or running setup
- `--non-interactive`
  Disable prompts even when stdin is a TTY
- `-y`, `--yes`
  Accept defaults and skip confirmations
- `-h`, `--help`
  Show help

## Initialization Modes

- `default`
  Use inferred strategies for detected ignored files and directories
- `custom`
  Choose a strategy for each detected ignored path
- `none`
  Skip reuse and setup entirely

## Reuse Strategies

- `copy`
  Copy a file or directory into the new worktree
- `symlink`
  Create a symlink to the source path
- `skip`
  Ignore the path completely
- `setup`
  Do not copy or symlink the path. Instead, run setup commands inside the new worktree

## Automatic Setup Detection

`wtx` detects common ecosystems and chooses setup commands automatically.

Examples:

- Node.js
  - `bun.lock` / `bun.lockb` -> `bun install`
  - `pnpm-lock.yaml` -> `pnpm install`
  - `yarn.lock` -> `yarn install`
  - `package.json` only -> `npm install`
- Python
  - `uv.lock` -> `uv sync`
  - `poetry.lock` -> `poetry install`
  - `Pipfile` / `Pipfile.lock` -> `pipenv install`
  - `requirements.txt` / `pyproject.toml` -> virtualenv + pip flow
- Go
  - `go.mod` -> `go mod download`
- Rust
  - `Cargo.toml` -> `cargo fetch`
- Java
  - `mvnw` + `pom.xml` -> `./mvnw dependency:resolve`
  - `pom.xml` -> `mvn dependency:resolve`
  - `gradlew` + `build.gradle` -> `./gradlew build`
  - `build.gradle` -> `gradle build`

## Config

`wtx` loads optional JSON config from:

- `~/.wtx.json`
- `<repo>/.wtx.json`

Project config overrides user config.

### Supported Fields

```json
{
  "worktreeRoot": "/Users/alex/worktrees",
  "strategyOverrides": {
    ".claude": "symlink",
    ".env": "copy",
    "node_modules": "setup"
  },
  "setupTemplates": {
    "node-pnpm": [
      {
        "id": "node-pnpm-frozen",
        "description": "Install Node.js dependencies with pnpm using the lockfile",
        "command": "pnpm",
        "args": ["install", "--frozen-lockfile"]
      }
    ]
  },
  "hooks": {
    "beforeCreate": [
      {
        "id": "announce-start",
        "description": "Announce worktree creation",
        "command": "echo",
        "args": ["before-create"]
      }
    ],
    "afterCreate": [
      {
        "id": "announce-finish",
        "description": "Announce worktree completion",
        "command": "echo",
        "args": ["after-create"]
      }
    ]
  }
}
```

### `worktreeRoot`

Sets the default parent directory for generated worktree paths.

### `strategyOverrides`

Overrides inferred strategies for specific ignored paths.

Example:

- `.env` -> `copy`
- `.claude` -> `symlink`
- `node_modules` -> `setup`

### `setupTemplates`

Replaces detected setup commands by template ID.

Example:

- Replace detected `node-pnpm` with `pnpm install --frozen-lockfile`
- Replace detected `python-uv-sync` with a custom `uv` invocation

### `hooks`

Runs additional commands before or after worktree creation.

Supported hook phases:

- `beforeCreate`
- `afterCreate`

Hook processes receive these environment variables:

- `WTX_REPO_ROOT`
- `WTX_WORKTREE_DIRECTORY`
- `WTX_PROJECT_NAME`
- `WTX_BRANCH_NAME`
- `WTX_BRANCH_MODE`

## Safety Notes

- `symlink` keeps the source and new worktree pointed at the same underlying files
- build outputs and transient directories are usually inferred as `skip`
- dependency environments like `node_modules` and `.venv` are better handled as `setup` than shared directly
- `--dry-run` is the safest way to inspect the full plan before any filesystem or Git changes happen

## Example Workflow

Create a new branch worktree with defaults:

```bash
wtx add feature/refactor-auth --new-branch
```

Preview everything without making changes:

```bash
wtx add feature/refactor-auth --new-branch --dry-run --non-interactive
```

Create multiple review worktrees under a shared root:

```bash
wtx batch-add review/a review/b review/c --new-branch --root ~/worktrees --non-interactive
```

## Project Structure

- [cmd/wtx/main.go](/Users/17admin/projects/wtx/cmd/wtx/main.go:1)
  CLI entrypoint
- [internal/cli/app.go](/Users/17admin/projects/wtx/internal/cli/app.go:1)
  argument parsing, prompts, execution flow
- [internal/core](/Users/17admin/projects/wtx/internal/core)
  defaults, planning, setup detection, types
- [internal/config/config.go](/Users/17admin/projects/wtx/internal/config/config.go:1)
  config loading and merge logic
- [internal/git](/Users/17admin/projects/wtx/internal/git)
  Git command construction and repository inspection
- [internal/fsops/apply.go](/Users/17admin/projects/wtx/internal/fsops/apply.go:1)
  copy, symlink, and setup execution
