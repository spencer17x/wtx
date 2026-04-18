const test = require("node:test");
const assert = require("node:assert/strict");

const {
  getAssetInfo,
  getVersionFromTag,
} = require("./platform");

test("maps darwin arm64 to the expected release asset", () => {
  assert.deepEqual(getAssetInfo("darwin", "arm64"), {
    goos: "Darwin",
    goarch: "arm64",
    archive: "wtx_Darwin_arm64.tar.gz",
    binaryName: "wtx",
  });
});

test("maps linux x64 to the expected release asset", () => {
  assert.deepEqual(getAssetInfo("linux", "x64"), {
    goos: "Linux",
    goarch: "x86_64",
    archive: "wtx_Linux_x86_64.tar.gz",
    binaryName: "wtx",
  });
});

test("throws for unsupported platforms", () => {
  assert.throws(
    () => getAssetInfo("win32", "x64"),
    /Unsupported platform: win32 x64/,
  );
});

test("strips the leading v from a release tag", () => {
  assert.equal(getVersionFromTag("v0.1.0"), "0.1.0");
});

test("rejects non-semver release tags", () => {
  assert.throws(
    () => getVersionFromTag("release-1"),
    /Expected a tag in the form vX.Y.Z/,
  );
});
