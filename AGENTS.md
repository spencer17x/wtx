# wtx Agent Guide

This file applies to the entire repository.

## Purpose

`wtx` is a **Go** CLI that wraps `git worktree add` with local environment setup (copy/symlink/setup of ignored files, project-type detection). The npm package `@spencer17x/wtx` is a **release/install helper**, not the product TypeScript app.

## Start Here

1. Read the user request and the relevant Go packages / `npm/` helpers.
2. Inspect `git status --short`; preserve unrelated work.
3. Prefer extending existing Go paths over inventing parallel Node logic.
4. **Do not** commit, push, tag, or publish releases unless the user asks.

## Repository Map

```text
cmd/ / internal/     Go CLI source (see go.mod)
npm/                 Node install/postinstall helpers for the published binary
.github/workflows/   Release automation
docs/                releasing and docs
```

## Coding Conventions

- Language packs: **go** + **tooling Node** (not typescript-node).
- Keep Go as the source of truth for CLI behavior.
- Keep the npm helper stack on the repository-pinned Node/npm toolchain; do not introduce pnpm without an intentional package-manager migration.

## Runtime And Environment

- Go: version from `go.mod`.
- Tooling Node/npm: `.nvmrc` and `packageManager` pin Node `24.18.0` and npm `11.16.0`; supported ranges come from `engines`.
- Install targets: macOS/Linux, arm64/x64.

## Commands And Verification

```bash
npm test
npm run release:check
```

| Change | Required checks |
| --- | --- |
| Go code | `npm test` |
| npm helpers | `npm test` |
| Release tooling | `npm run release:check`; follow `docs/releasing.md` only when asked |

## Git And Commits

- Conventional Commits core types: `feat` `fix` `docs` `refactor` `perf` `test` `build` `ci` `chore` `revert`.
- Header ≤100; no trailing period.

## Security And Privacy

- Do not commit secrets or machine-local paths into fixtures.
- Release credentials only via GitHub Actions secrets.

## Definition Of Done

- [ ] Behavior complete
- [ ] Relevant tests green or skips disclosed
- [ ] Handoff lists verification and residual risk

## Repository Engineering Baseline

- This repository is self-contained; it does not depend on a shared standards repository.
- Go version comes from `go.mod`; tooling Node comes from `.nvmrc` and `engines`.
- No pre-commit, commit-msg, or pre-push hooks are configured.
- Pull requests and branch pushes do not run automated CI or lint checks.
- Release workflows run only when manually requested or when a `v*` tag is pushed.
