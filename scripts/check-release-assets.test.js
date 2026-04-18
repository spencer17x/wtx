const test = require("node:test");
const assert = require("node:assert/strict");

const {
  assertReleaseAssetsComplete,
  getRequiredReleaseAssetNames,
  validateCliArgs,
} = require("./check-release-assets");

function buildRelease(assetOverrides = {}) {
  const overrideEntries = new Map(Object.entries(assetOverrides));

  return {
    assets: getRequiredReleaseAssetNames().map((name) => ({
      name,
      size: 1,
      ...overrideEntries.get(name),
    })),
  };
}

test("lists the required release archives and checksum manifest", () => {
  assert.deepEqual(getRequiredReleaseAssetNames(), [
    "wtx_Darwin_arm64.tar.gz",
    "wtx_Darwin_x86_64.tar.gz",
    "wtx_Linux_arm64.tar.gz",
    "wtx_Linux_x86_64.tar.gz",
    "checksums.txt",
  ]);
});

test("accepts a release that contains every required asset", () => {
  assert.doesNotThrow(() => assertReleaseAssetsComplete(buildRelease()));
});

test("rejects a release that is missing required assets", () => {
  const release = buildRelease();
  release.assets = release.assets.filter((asset) => asset.name !== "checksums.txt");

  assert.throws(
    () => assertReleaseAssetsComplete(release),
    /Release assets are incomplete\. Missing: checksums\.txt/,
  );
});

test("rejects a release with an empty required asset", () => {
  assert.throws(
    () =>
      assertReleaseAssetsComplete(
        buildRelease({
          "wtx_Linux_x86_64.tar.gz": { size: 0 },
        }),
      ),
    /Release assets are incomplete\. Empty: wtx_Linux_x86_64\.tar\.gz/,
  );
});

test("rejects missing cli arguments", () => {
  assert.throws(
    () => validateCliArgs([]),
    /Usage: node scripts\/check-release-assets\.js <releaseJsonPath>/,
  );
});
