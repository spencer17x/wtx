# Packaging and Release Design

## Goal

Add first-class distribution for `wtx` through:

- `npm install -g wtx`
- Homebrew via a dedicated tap, so users can run `brew tap spencer17x/wtx` once and then `brew install wtx`
- automated GitHub Releases with published binaries

This design does not assume acceptance into `homebrew/core`. The project should remain compatible with a future `homebrew/core` submission, but that external review process is out of scope for the implementation.

## Current Context

`wtx` is currently a Go CLI with no packaging metadata, no GitHub Actions workflows, no npm package definition, no release automation, and no Homebrew formula repository automation.

The current repository contains:

- Go CLI entrypoint in `cmd/wtx/main.go`
- internal packages under `internal/`
- English and Chinese READMEs

The repository currently has no GitHub Releases and no packaging repos or workflow files.

## Scope

### In Scope

- Add an npm package named `wtx` from this repository
- Add automated binary release artifacts for supported platforms
- Add tag-driven GitHub Release automation
- Add automation that updates a dedicated Homebrew tap repository
- Document how and when releases are created
- Document required secrets and external setup
- Add CI checks for packaging-related behavior

### Out of Scope

- Automatic submission to `homebrew/core`
- Windows distribution
- Package managers beyond npm and Homebrew
- Automatic semantic version calculation from commit history

## Distribution Model

### npm

The root repository becomes an npm package named `wtx`.

The npm package is a thin installer wrapper, not a JavaScript implementation of the CLI. On installation, it determines the current OS and CPU architecture, downloads the matching prebuilt binary from the GitHub Release assets for the package version, stores it inside the package, and exposes it through the npm `bin` entry.

This keeps npm users on the same binary artifacts as GitHub Releases and Homebrew users.

### Homebrew

Homebrew distribution uses a custom tap repository, `spencer17x/homebrew-wtx`.

Users install it through either:

- `brew tap spencer17x/wtx && brew install wtx`
- `brew install spencer17x/wtx/wtx`

The implementation in this repository should update the tap automatically on tagged releases by committing a new `Formula/wtx.rb` with the current version and source tarball checksum.

### GitHub Releases

GitHub Releases are the source of truth for shipped binary artifacts.

Each tagged version should publish:

- archives for supported OS and architecture pairs
- a checksum file for all assets
- release notes generated or attached by the workflow

## Supported Platforms

Initial binary targets:

- macOS arm64
- macOS amd64
- Linux amd64
- Linux arm64

Archive naming should be deterministic and easy for both npm install scripts and Homebrew tooling to consume. A consistent convention such as the following is recommended:

- `wtx_Darwin_arm64.tar.gz`
- `wtx_Darwin_x86_64.tar.gz`
- `wtx_Linux_x86_64.tar.gz`
- `wtx_Linux_arm64.tar.gz`
- `checksums.txt`

## Release Policy

### Normal Pushes and Pull Requests

Pushes to branches and pull requests run verification only. They do not publish any package, create any GitHub Release, or modify the Homebrew tap.

Verification includes:

- `go test ./...`
- Go build smoke checks
- npm packaging validation for the installer wrapper
- release configuration validation where possible

### Tagged Releases

Publishing happens only from semantic version tags in the form `vX.Y.Z`.

Example:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Pushing such a tag triggers the release workflow, which must:

1. verify tests
2. build release artifacts
3. create or update the GitHub Release
4. publish the npm package `wtx`
5. update the Homebrew tap formula to the same version

### Prereleases

Prerelease tags such as `v0.1.0-rc.1` are optional future support. This design does not require implementing prerelease handling in the first pass.

## Required Repository Changes

### Packaging Metadata

- Add `package.json` at the repository root
- Add a package lockfile if the npm workflow needs one
- Add npm installer scripts under a dedicated packaging directory

The npm package should include:

- package name `wtx`
- a version field that the release workflow updates from the Git tag before publishing
- `bin` entry mapping `wtx` to the installed binary launcher path
- metadata pointing to the GitHub repository

### Release Tooling

- Add `.goreleaser.yaml` to define build targets, archives, checksums, and release publishing
- Add scripts to support npm binary download/install
- Add scripts to update the Homebrew formula from release metadata

### GitHub Actions

- Add `.github/workflows/ci.yml`
- Add `.github/workflows/release.yml`
- Add `.github/workflows/release-dry-run.yml`

### Documentation

- Update `README.md`
- Update `README.zh-CN.md`
- Add `docs/releasing.md`

## Workflow Design

### CI Workflow

Trigger:

- push
- pull_request

Responsibilities:

- run `go test ./...`
- run a normal build of the CLI
- verify npm package metadata is valid
- run packaging/unit tests for the installer scripts
- optionally run a GoReleaser validation or snapshot build

The CI workflow must not publish anything.

### Release Workflow

Trigger:

- push of tags matching `v*`

Responsibilities:

- check out the repository
- set up Go and Node.js
- run tests again for release safety
- derive the release version from the pushed tag and apply it to npm package metadata
- run GoReleaser to build archives, checksums, and GitHub Release assets
- publish the npm package using `NPM_TOKEN`
- update the Homebrew tap repo using a repo-scoped token

### Release Dry Run Workflow

Trigger:

- `workflow_dispatch`

Responsibilities:

- run the release pipeline in snapshot or validation mode
- produce artifacts for inspection
- skip npm publish and tap updates

This workflow exists to validate packaging changes safely before cutting a real tag.

## Secret and External Dependency Model

### Secrets

The release workflow requires at least:

- `NPM_TOKEN`
- `HOMEBREW_TAP_TOKEN`

`NPM_TOKEN` must have publish access to the npm package name `wtx`.

`HOMEBREW_TAP_TOKEN` must have permission to push commits to `spencer17x/homebrew-wtx`.

### External Repositories

The following repositories/accounts must exist before the full release workflow can succeed:

- npm package ownership for `wtx`
- tap repository `spencer17x/homebrew-wtx`

### GitHub Permissions

The GitHub Actions workflow should use the minimum required permissions. The GitHub-provided token can create release assets if configured with `contents: write`, but a dedicated token is still required for pushing changes to a separate tap repository.

## Homebrew Formula Strategy

The tap formula should build from the GitHub source tarball for the tagged release version.

The formula should:

- declare the project metadata
- depend on Go for building from source
- build `./cmd/wtx`
- include a minimal install test such as `wtx --help`

The formula file should be rendered from a template or from a script that updates:

- version
- source tarball URL
- SHA256 checksum

The source tarball checksum is sufficient for a build-from-source formula in the tap.

## npm Installer Strategy

The npm package should install the matching binary for the current platform.

Key requirements:

- deterministic platform mapping from Node.js `process.platform` and `process.arch`
- clear errors for unsupported platforms
- executable permissions set correctly after download and extraction
- `wtx` command available through npm global install
- version-aligned download URLs using the npm package version

The installer must fail loudly and clearly if:

- the expected release asset does not exist
- the current platform is unsupported
- the downloaded archive checksum cannot be validated, if checksum validation is implemented in the first pass

## Failure Behavior

GitHub Release artifacts are the primary output. npm and Homebrew are downstream.

Expected failure behavior:

- If tests fail, nothing publishes
- If archive build fails, nothing publishes
- If GitHub Release creation fails, nothing publishes downstream
- If npm publish fails after assets are created, the workflow fails and can be retried after fixing npm state or credentials
- If Homebrew tap update fails after npm succeeds, the workflow fails and requires rerun or manual repair of the tap

This asymmetry is acceptable because the release assets remain available and authoritative.

## Testing Strategy

### Go Tests

Existing Go tests continue to validate core CLI behavior.

### Packaging Tests

Add tests for:

- platform-to-asset mapping logic in the npm installer
- release metadata rendering logic for the Homebrew formula updater

These tests should cover unsupported platforms and naming mismatches.

### Workflow Validation

The repository should support at least one local or CI-level dry-run path for release validation, such as:

- GoReleaser check or snapshot mode
- a manual GitHub Actions dry-run workflow

## File Plan

Expected new files:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/release-dry-run.yml`
- `.goreleaser.yaml`
- `package.json`
- `npm/install.js`
- `npm/platform.js`
- `scripts/update-homebrew-formula.js`
- `packaging/homebrew/wtx.rb.tmpl`
- `docs/releasing.md`

Expected modified files:

- `README.md`
- `README.zh-CN.md`

Optional test files, depending on implementation language for scripts:

- `npm/platform.test.js`
- `scripts/update-homebrew-formula.test.js`

## Recommended Implementation Sequence

1. Add packaging tests and npm platform mapping logic
2. Add npm package metadata and local install wrapper behavior
3. Add GoReleaser configuration and snapshot validation
4. Add Homebrew formula rendering/update logic
5. Add CI workflow
6. Add release workflow
7. Add release dry-run workflow
8. Add release documentation and README install sections

## Success Criteria

The design is successful when:

- a tagged release creates GitHub Release assets automatically
- `npm install -g wtx` installs a working `wtx` command from release binaries
- the Homebrew tap formula updates automatically on release
- users can run `brew tap spencer17x/wtx` and then `brew install wtx`
- normal pushes and pull requests never publish by accident
- release steps and required secrets are documented clearly enough to operate without guesswork
