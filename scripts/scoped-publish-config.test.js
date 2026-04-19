const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");

test("package metadata and release workflow are configured for scoped npm publish", () => {
  const pkg = JSON.parse(fs.readFileSync("package.json", "utf8"));
  const workflow = fs.readFileSync(".github/workflows/release.yml", "utf8");

  assert.equal(pkg.name, "@spencer17x/wtx");
  assert.equal(pkg.bin.wtx, "npm/bin/wtx.js");
  assert.match(workflow, /run: npm publish --access public/);
});
