# Releasing wtx

## Required Secrets

- `NPM_TOKEN`
- `HOMEBREW_TAP_TOKEN`

## Required External Setup

- npm package ownership for `wtx`
- tap repository `spencer17x/homebrew-wtx`

## Normal Development

Branch pushes and pull requests run CI only.

## Shipping a Release

1. Verify `main` is green.
2. Create a tag in the form `vX.Y.Z`.
3. Push the tag.

```bash
git tag v0.1.0
git push origin v0.1.0
```

That tag triggers:

- GitHub Release asset publishing
- npm publish
- Homebrew tap update
