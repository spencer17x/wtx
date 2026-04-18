const test = require("node:test");
const assert = require("node:assert/strict");

const {
  getAssetInfo,
  getReleaseArchives,
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

test("maps darwin x64 to the expected release asset", () => {
  assert.deepEqual(getAssetInfo("darwin", "x64"), {
    goos: "Darwin",
    goarch: "x86_64",
    archive: "wtx_Darwin_x86_64.tar.gz",
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

test("maps linux arm64 to the expected release asset", () => {
  assert.deepEqual(getAssetInfo("linux", "arm64"), {
    goos: "Linux",
    goarch: "arm64",
    archive: "wtx_Linux_arm64.tar.gz",
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

test("rejects release tags with leading zeroes", () => {
  assert.throws(
    () => getVersionFromTag("v01.2.3"),
    /Expected a tag in the form vX.Y.Z/,
  );
});

test("lists the release archives used by installers and release checks", () => {
  assert.deepEqual(getReleaseArchives(), [
    "wtx_Darwin_arm64.tar.gz",
    "wtx_Darwin_x86_64.tar.gz",
    "wtx_Linux_arm64.tar.gz",
    "wtx_Linux_x86_64.tar.gz",
  ]);
});
