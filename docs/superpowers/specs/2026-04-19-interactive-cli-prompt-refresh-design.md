# Interactive CLI Prompt Refresh Design

## Goal

Improve the interactive `wtx add` experience so that:

- selection prompts use real terminal selection behavior instead of typed numeric options
- the worktree directory is presented as an explicit, editable input step during interactive flow
- interactive UX becomes easier to understand without changing existing command semantics

## Current Context

`wtx` already supports interactive prompting in `internal/cli/app.go`, but the current prompt layer is intentionally minimal:

- `askSelect` prints numbered options and expects the user to type a number
- project name and directory are grouped under a secondary edit menu
- the default directory is derived from the branch or project name and sanitized for filesystem safety
- non-interactive behavior is controlled by `--non-interactive` and `--yes`

This creates two practical UX problems:

1. “selection” prompts do not feel like real selection controls
2. worktree directory editing is technically available, but discoverability is poor and the flow makes it easy to think custom directory entry is unsupported

## User-Facing Requirements

### In Scope

- interactive branch mode selection should support arrow-key style choice and Enter confirmation
- interactive initialization mode selection should support the same selection behavior
- per-path strategy selection in custom initialization mode should use the same selection behavior
- interactive flow should explicitly ask for project name
- interactive flow should explicitly ask for worktree directory
- the directory prompt should clearly show the computed default value
- users must be able to replace the default directory with any custom path

### Out of Scope

- changing non-interactive CLI flags or semantics
- changing batch creation behavior beyond prompt plumbing reuse where safe
- changing how default directories are sanitized
- changing branch creation semantics

## Design Summary

The interactive prompt layer will be upgraded from a handwritten numeric menu UI to a small terminal prompt library, while keeping the existing `promptUI` abstraction as the application-facing boundary.

The interactive `add` flow will also be reshaped so that project name and worktree directory are gathered as explicit sequential inputs instead of being hidden behind a “choose what to edit” submenu.

## Prompt Library Strategy

### Recommended Approach

Introduce `github.com/manifoldco/promptui` as the interactive prompt dependency. It will provide:

- single-select prompts with arrow-key navigation
- text input with default values
- yes/no confirmation prompts

The rest of the CLI should continue talking to `promptUI` methods such as:

- `askSelect`
- `askInput`
- `askConfirm`

This keeps the third-party dependency isolated to the prompt implementation layer and avoids spreading library-specific concepts throughout the planner and execution logic.

### Why This Approach

- it directly solves the typed-number selection problem
- it minimizes changes to the rest of the codebase
- it preserves non-interactive flows because prompt usage remains gated by `shouldUsePrompts`
- it gives the project a more professional terminal UX without redesigning the command model

## Interactive Flow Changes

### Current Flow

Today, project name and directory are shown together and then hidden behind a follow-up menu:

1. show current project name and directory
2. ask whether to accept values or edit project name or directory or both
3. return to the summary screen until accepted

This is compact, but not very discoverable.

### New Flow

In interactive `add`, the flow should become:

1. choose branch mode with interactive selection
2. enter or confirm branch name
3. if creating a new branch, enter or confirm base ref
4. enter or confirm project name
5. enter or confirm worktree directory
6. choose initialization mode with interactive selection
7. if custom mode, choose a strategy per ignored path with interactive selection
8. review the final plan
9. confirm creation

### Directory Prompt Behavior

The directory prompt should always be an explicit step in interactive mode.

The default shown in that step should be:

- the CLI-provided `--dir` value if present
- otherwise the derived default built from the effective project name and worktree root

If the user edits the project name first, the suggested default directory should be recalculated before asking for the directory. If the user then enters a custom directory, that custom value must be preserved as the final directory.

This removes the current ambiguity where editing the project name can rebuild the derived directory later in the flow.

## Behavioral Rules

### Interactive Mode

- interactive prompts only appear when stdin is a TTY and neither `--yes` nor `--non-interactive` was supplied
- selection prompts should never ask the user to type numeric option ids
- text prompts should continue to accept Enter for default values

### Non-Interactive Mode

Non-interactive behavior remains unchanged:

- `--dir` still overrides the directory directly
- `--root` still controls the derived parent directory
- `--yes` still accepts defaults without prompts
- `--non-interactive` still disables prompts

### Batch Mode

`batch-add` remains non-interactive. It should not inherit the new interactive wizard behavior. Existing restrictions such as no `--dir` in batch mode stay unchanged.

## Implementation Shape

### `promptUI`

`promptUI` remains the boundary between CLI flow logic and terminal interaction. Its methods will keep their existing meanings, but their implementations will be backed by `promptui` for interactive use.

This design intentionally avoids rewriting the rest of the app to know about the external prompt dependency.

### `resolveProjectNameAndDirectory`

This function should be simplified from a menu loop into a sequential input flow:

1. compute defaults
2. prompt for project name
3. recompute the derived directory suggestion from the effective project name if no explicit directory flag was given
4. prompt for directory using that effective default
5. normalize relative paths to absolute paths exactly as today

This is the key structural change for the directory UX issue.

## Error Handling

- if the prompt library returns an interrupt or terminal read error, the command should return that error without partial execution
- invalid non-interactive combinations should continue to fail fast as they do today
- prompt failures must happen before any Git or filesystem mutation

## Testing Strategy

### Existing Coverage to Preserve

- argv parsing behavior
- prompt gating via `shouldUsePrompts`
- dry-run execution behavior

### New Coverage to Add

- interactive selection wrapper returns the selected label correctly
- project name prompt is asked explicitly in interactive mode
- directory prompt is asked explicitly in interactive mode
- changing project name updates the suggested default directory
- entering a custom directory preserves that directory as the final result
- non-interactive directory derivation remains unchanged

### Testability Design

Because real terminal widgets are harder to exercise in unit tests, prompt behavior should be isolated behind testable seams. Tests should be able to stub prompt responses without requiring a live TTY UI.

That means the implementation should avoid burying application logic directly inside library-specific prompt callbacks.

## Risks

### Dependency Risk

Adding `promptui` introduces the project’s first external Go dependency. This is acceptable for the UX win, but the dependency should remain narrow in scope and easy to replace if needed.

### Terminal Compatibility Risk

Interactive terminal libraries can behave differently across shells and environments. Non-interactive mode must remain fully functional so automation is not impacted if an interactive environment is unusual.

### Flow Regression Risk

Changing the prompt sequence may unintentionally alter existing defaults. Tests should focus on preserving current non-interactive behavior and current directory derivation rules.

## Success Criteria

This design is successful when:

- interactive selection prompts no longer require typing numeric menu options
- interactive users are explicitly asked for the worktree directory
- users can enter a custom worktree directory in interactive mode without confusion
- non-interactive CLI behavior remains unchanged
- existing tests still pass and new prompt-flow tests cover the new behavior
