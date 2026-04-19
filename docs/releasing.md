# Releasing wtx

## Required Secrets

- `NPM_TOKEN` with npm publish access

## Required External Setup

- npm package ownership for `wtx`

## Normal Development

Non-tag branch pushes and pull requests run CI only. Tag pushes still run CI, and they also trigger the release workflow.

## Release Dry Run

Use the `Release Dry Run` workflow before pushing the real tag when you want to validate a candidate release from an arbitrary branch, tag, or commit.

- `ref`: the branch name, existing tag, or commit SHA to check out
- `version`: the candidate release version in `vX.Y.Z` form

The dry run validates the candidate version, runs `npm test`, runs `npm run release:check`, checks `npm pack --dry-run`, and runs `npm run release:snapshot`. It does not publish anything.

## Shipping a Release

1. Verify `main` is green.
2. Optionally run `Release Dry Run` against the exact ref and candidate version you plan to release.
3. Create a tag in the form `vX.Y.Z` on the commit being released, typically the intended `main` tip after CI is green.
4. Push the tag.

```bash
git tag v0.1.0
git push origin v0.1.0
```

That tag triggers:

- GitHub Release asset publishing
- npm publish

The workflow keeps the git tree clean while GoReleaser runs. It derives the npm package version from the tag only after the GitHub Release step, right before npm publish.

## Rerun Behavior

GitHub Release asset completeness gates npm publishing on reruns.

- If the release does not exist yet, the workflow runs GoReleaser to create the release and upload assets.
- If the release already exists and contains the expected archives plus `checksums.txt`, the workflow skips GoReleaser and can continue with downstream publish steps.
- If the release object exists but any required asset is missing or empty, the workflow fails closed instead of publishing npm from a partial release.
