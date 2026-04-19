# Worktree Navigation Design

**Date:** 2026-04-19

**Goal:** Add branch-based worktree navigation to `wtx` without pretending a standalone CLI binary can change the current shell's working directory on its own.

## Summary

`wtx` will remain an enhancement layer around high-frequency worktree workflows, not a full compatibility wrapper for every `git worktree` subcommand. This design adds a small navigation surface:

- `wtx path <branch>`
- `wtx switch <branch>`
- `wtx shell-init <zsh|bash>`

The new commands operate only on exact branch names. They resolve worktree locations from Git's worktree metadata rather than guessing from directory names. Shell integration stays optional. Without shell integration, `switch` behaves as a friendly navigation helper that prints the resolved path and tells the user how to `cd` into it. With shell integration enabled, `wtx switch <branch>` becomes a real shell-level directory jump.

## Product Scope

`wtx` does not aim to fully mirror `git worktree`. It keeps its current role as a focused developer workflow tool:

- `git worktree` remains the source of truth for low-level, general-purpose worktree management
- `wtx` continues to focus on higher-level flows such as creation, initialization, and now navigation

This feature intentionally does not add full `git worktree` command compatibility. The supported surface remains explicit and small.

## Command Surface

### `wtx path <branch>`

Returns the absolute path for the worktree currently associated with the exact branch name.

Behavior:

- Accepts only one positional argument: the branch name
- Requires an exact branch name such as `feature/my-branch`
- Outputs only the resolved absolute path on success
- Is safe for scripting and shell composition, for example:

```bash
cd "$(wtx path feature/my-branch)"
```

Non-goals:

- No project-name lookup
- No directory-name lookup
- No fuzzy matching
- No path input

### `wtx switch <branch>`

Provides a user-facing navigation command that shares the same branch resolution logic as `wtx path`.

Behavior without shell integration:

- Resolves the target worktree using the same exact branch rules as `wtx path`
- Prints the resolved absolute path
- Prints a short follow-up instruction telling the user how to jump with shell integration disabled, for example:

```text
/Users/alex/projects/feature-my-branch
shell integration not enabled; use: cd "$(wtx path feature/my-branch)"
```

Behavior with shell integration enabled:

- The shell wrapper intercepts `wtx switch <branch>`
- The wrapper runs `command wtx path <branch>` and executes `cd` in the current shell
- The Go binary itself never claims to mutate the current shell session directly

### `wtx shell-init <shell>`

Prints shell integration code for supported shells.

Initial support:

- `wtx shell-init zsh`
- `wtx shell-init bash`

Usage example:

```bash
eval "$(wtx shell-init zsh)"
```

The generated shell function:

- Forwards all non-`switch` subcommands to the real `wtx` binary
- Rewrites `wtx switch <branch>` into shell-native `cd "$(command wtx path <branch>)"`
- Keeps shell integration optional rather than making it part of installation requirements

## Resolution Rules

Both `path` and `switch` use a shared branch-to-worktree resolution implementation.

Rules:

- Match only exact branch names supplied by the user
- Resolve against Git worktree metadata, not folder names
- Match only local branch refs in the form `refs/heads/<branch>`
- Ignore detached-HEAD worktrees

This keeps the first release unambiguous and avoids collisions such as:

- branch name: `feature/foo`
- sanitized directory name: `feature-foo`

The CLI must not guess between them.

## Prunable Worktrees

`prunable` worktrees are treated as invalid navigation targets.

Behavior:

- If the branch resolves to a prunable worktree record, `wtx path` and `wtx switch` fail
- The error should explicitly tell the user what to do next

Recommended error text:

```text
worktree for branch feature/my-branch is prunable; run git worktree prune
```

This avoids "successful" navigation to stale metadata.

## Error Handling

The commands should produce explicit, actionable errors for these cases:

- The current directory is not inside a Git repository
- The branch has no associated worktree in the current repository
- The resolved worktree record is marked prunable
- `wtx shell-init` receives an unsupported shell value

Recommended error shape:

- `no worktree found for branch: feature/my-branch`
- `worktree for branch feature/my-branch is prunable; run git worktree prune`
- `unsupported shell: fish`

## Architecture

Implementation should stay close to the current code layout and avoid over-expanding the CLI surface.

Suggested structure:

- Extend CLI parsing in `internal/cli/app.go` to recognize `path`, `switch`, and `shell-init`
- Add a small worktree inspection helper under `internal/git` to enumerate repository worktrees and their associated refs
- Add a shared resolver that converts `branch -> absolute path`
- Keep shell-init generation in the CLI layer because it is user-facing presentation logic

The key boundary is:

- `internal/git` knows how to inspect Git worktree state
- `internal/cli` decides how commands present that state to the user

## Testing Strategy

Add focused unit coverage for:

- `wtx path` resolving an exact branch to an absolute path
- `wtx path` returning a clear error when no worktree exists for the branch
- `wtx path` returning a clear error for prunable worktrees
- `wtx switch` printing path plus shell-integration guidance when no wrapper is active
- `wtx shell-init zsh` and `wtx shell-init bash` generating wrappers that rewrite `switch` into `cd "$(command wtx path ...)"` while forwarding all other subcommands unchanged

Also preserve existing command behavior for:

- `add`
- `batch-add`
- help output

## Out of Scope

This design intentionally excludes:

- Full `git worktree` compatibility
- Project-name based navigation
- Directory-name based navigation
- Fuzzy or prefix branch matching
- Path-based navigation targets
- Additional shells beyond `zsh` and `bash`
- Automatic pruning or cleanup of stale worktrees

## Recommendation

Implement the recommended hybrid model:

1. Add a script-friendly `wtx path <branch>`
2. Add a user-facing `wtx switch <branch>` with graceful fallback behavior
3. Add optional `wtx shell-init <shell>` integration so users can get true in-shell jumps without making shell setup mandatory

This gives `wtx` a clean navigation story without stretching the CLI beyond what a standalone process can honestly do.
