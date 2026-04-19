# Packaging and Release Design

## Goal

Add first-class distribution for `wtx` through:

- `npm install -g wtx`
- automated GitHub Releases with published binaries

## Current Context

`wtx` is a Go CLI that ships from this repository and needs packaging and release automation centered on GitHub Releases plus npm distribution.

The repository contains:

- Go CLI entrypoint in `cmd/wtx/main.go`
- internal packages under `internal/`
- English and Chinese READMEs
- packaging and release workflow files

## Scope

### In Scope

- Ship an npm package named `wtx` from this repository
- Publish automated binary release artifacts for supported platforms
- Run tag-driven GitHub Release automation
- Document how releases are created and what credentials they need
- Add CI checks for packaging behavior

### Out of Scope

- Windows distribution
- Package managers beyond npm
- Automatic semantic version calculation from commit history

## Distribution Model

### npm

The root repository becomes an npm package named `wtx`.

The npm package is a thin installer wrapper rather than a JavaScript implementation of the CLI. On installation, it determines the current OS and CPU architecture, downloads the matching prebuilt binary from the GitHub Release assets for the package version, stores it inside the package, and exposes it through the npm `bin` entry.

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

Archive naming should be deterministic and easy for the npm installer tooling to consume. A consistent convention such as the following is recommended:

- `wtx_Darwin_arm64.tar.gz`
- `wtx_Darwin_x86_64.tar.gz`
- `wtx_Linux_x86_64.tar.gz`
- `wtx_Linux_arm64.tar.gz`
- `checksums.txt`

## Release Policy

### Normal Pushes and Pull Requests

Pushes to branches and pull requests run verification only. They do not publish any package or create any GitHub Release.

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

### Prereleases

Prerelease tags such as `v0.1.0-rc.1` are optional future support. This design does not require implementing prerelease handling in the first pass.

## Required Repository Changes

### Packaging Metadata

- Add `package.json` at the repository root
- Add npm installer scripts under a dedicated packaging directory

The npm package should include:

- package name `wtx`
- a version field that the release workflow updates from the Git tag before publishing
- `bin` entry mapping `wtx` to the installed binary launcher path
- metadata pointing to the GitHub repository

### Release Tooling

- Add `.goreleaser.yaml` to define build targets, archives, checksums, and release publishing
- Add scripts to support npm binary download and install

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
- run packaging and unit tests for the npm installer scripts
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

### Release Dry Run Workflow

Trigger:

- `workflow_dispatch`

Responsibilities:

- run the release pipeline in snapshot or validation mode
- produce artifacts for inspection
- skip npm publish

This workflow exists to validate packaging changes safely before cutting a real tag.

## Secret and External Dependency Model

### Secrets

The release workflow requires:

- `NPM_TOKEN`

`NPM_TOKEN` must have publish access to the npm package name `wtx`.

### External Accounts

The following must exist before the full release workflow can succeed:

- npm package ownership for `wtx`

### GitHub Permissions

The GitHub Actions workflow should use the minimum required permissions. The GitHub-provided token can create release assets if configured with `contents: write`, and a dedicated npm token handles package publication.

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
- the downloaded archive cannot be extracted

## Failure Behavior

GitHub Release artifacts are the primary output. npm is downstream.

Expected failure behavior:

- If tests fail, nothing publishes
- If archive build fails, nothing publishes
- If GitHub Release creation fails, nothing publishes downstream
- If npm publish fails after assets are created, the workflow fails and can be retried after fixing npm state or credentials

## Testing Strategy

### Go Tests

Existing Go tests continue to validate core CLI behavior.

### Packaging Tests

Add tests for:

- platform-to-asset mapping logic in the npm installer

These tests should cover unsupported platforms and naming mismatches.

### Workflow Validation

The repository should support at least one local or CI-level dry-run path for release validation, such as:

- GoReleaser check or snapshot mode
- a manual GitHub Actions dry-run workflow

## File Plan

Expected files:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/release-dry-run.yml`
- `.goreleaser.yaml`
- `package.json`
- `npm/install.js`
- `npm/platform.js`
- `docs/releasing.md`

Expected modified files:

- `README.md`
- `README.zh-CN.md`

Optional test files, depending on implementation language for scripts:

- `npm/platform.test.js`

## Recommended Implementation Sequence

1. Add packaging tests and npm platform mapping logic
2. Add npm package metadata and local install wrapper behavior
3. Add GoReleaser configuration and snapshot validation
4. Add CI workflow
5. Add release workflow
6. Add release dry-run workflow
7. Add release documentation and README install sections

## Success Criteria

The design is successful when:

- a tagged release creates GitHub Release assets automatically
- `npm install -g wtx` installs a working `wtx` command from release binaries
- normal pushes and pull requests never publish by accident
- release steps and required secrets are documented clearly enough to operate without guesswork
