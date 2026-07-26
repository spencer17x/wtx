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
.github/workflows/   CI (gofmt, go vet, npm tests, release)
docs/                releasing and docs
```

## Coding Conventions

- Language packs: **go** + **tooling Node** (not typescript-node).
- Keep Go as the source of truth for CLI behavior.
- Keep the npm helper stack on the repository-pinned Node/npm toolchain; do not introduce pnpm without an intentional package-manager migration.

## Runtime And Environment

- Go: version from `go.mod` / CI.
- Tooling Node/npm: `.nvmrc` and `packageManager` pin Node `24.18.0` and npm `11.16.0`; supported ranges come from `engines`.
- Install targets: macOS/Linux, arm64/x64.

## Commands And Verification

```bash
npm run check
```

| Change | Required checks |
| --- | --- |
| Go code | `npm run check` |
| npm helpers | `npm run check` |
| Release tooling | follow `docs/releasing.md` only when asked |

## Git And Commits

- Conventional Commits core types: `feat` `fix` `docs` `refactor` `perf` `test` `build` `ci` `chore` `revert`.
- Header ≤100; no trailing period.

## Security And Privacy

- Do not commit secrets or machine-local paths into fixtures.
- Release credentials only via CI secrets.

## Definition Of Done

- [ ] Behavior complete
- [ ] Go/npm checks green or skips disclosed
- [ ] Handoff lists verification and residual risk

## Repository Engineering Baseline

- This repository is self-contained; it does not depend on a shared standards repository.
- Go version comes from `go.mod`; tooling Node comes from `.nvmrc` and `engines`.
- `npm run check` is the single local and CI quality gate.
- Hooks are intentionally omitted. CI enforces the same gate on every push and pull request.
- Extend the existing `.github/workflows/ci.yml`; do not add a parallel quality workflow.
