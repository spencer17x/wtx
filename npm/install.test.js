const test = require("node:test");
const assert = require("node:assert/strict");

const {
  buildReleaseAssetUrl,
  resolveBinaryPath,
} = require("./install");

test("builds the expected GitHub release asset URL", () => {
  assert.equal(
    buildReleaseAssetUrl({
      owner: "spencer17x",
      repo: "wtx",
      version: "0.1.0",
      archive: "wtx_Darwin_arm64.tar.gz",
    }),
    "https://github.com/spencer17x/wtx/releases/download/v0.1.0/wtx_Darwin_arm64.tar.gz",
  );
});

test("resolves the local binary path inside the npm package", () => {
  assert.match(
    resolveBinaryPath("/tmp/pkg", "wtx"),
    /\/tmp\/pkg\/npm\/bin\/wtx$/,
  );
});
