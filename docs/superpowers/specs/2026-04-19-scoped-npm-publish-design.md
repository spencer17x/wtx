# Scoped npm Publish Design

## Summary

Move the npm distribution for `wtx` from the blocked unscoped package name `wtx` to the scoped package name `@spencer17x/wtx`.

This keeps the installed CLI command name as `wtx`, while making npm publishing comply with npm package naming policy and the registry error currently blocking release.

## Goals

- Publish the package to npm under `@spencer17x/wtx`
- Keep the installed executable name as `wtx`
- Update release automation so scoped publishes use public visibility
- Update user-facing docs so install and release instructions match the new package name

## Non-Goals

- Renaming the GitHub repository
- Renaming the CLI executable from `wtx`
- Changing GitHub Release asset names
- Changing Homebrew or adding any new distribution channel

## Design

### Package Identity

- Change `package.json` `name` from `wtx` to `@spencer17x/wtx`
- Keep `bin.wtx` unchanged so global installs still expose the command `wtx`
- Keep the development versioning model unchanged: tags still drive the published semver

### Release Workflow

- Keep the current tag-triggered GitHub Actions release flow
- Keep GoReleaser responsible for GitHub Release assets
- Keep npm version derivation from the `vX.Y.Z` tag
- Change npm publish to `npm publish --access public`, because scoped packages are private by default on npm unless published as public
- Keep the existing npm package existence check, but ensure it resolves the scoped package name from `package.json`

### Documentation

- Update README install instructions to use `npm install -g @spencer17x/wtx`
- Explicitly document that the npm package name is scoped, but the installed command remains `wtx`
- Update release docs to describe npm ownership and publish setup for `@spencer17x/wtx`

## Validation

- `npm test`
- `npm run release:check`
- Verify release workflow syntax still matches the intended flow after changing the publish command

## Risks and Mitigations

- Existing users may expect `npm install -g wtx`
  - Mitigation: document the new install command clearly in both READMEs
- Scoped publishes default to private
  - Mitigation: make the workflow explicitly publish with `--access public`
