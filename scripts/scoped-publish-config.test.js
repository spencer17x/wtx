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

test("scoped npm package documentation keeps the CLI command as wtx", () => {
  const readme = fs.readFileSync("README.md", "utf8");
  const readmeZh = fs.readFileSync("README.zh-CN.md", "utf8");
  const releasing = fs.readFileSync("docs/releasing.md", "utf8");
  const releasingZh = fs.readFileSync("docs/releasing.zh-CN.md", "utf8");

  assert.match(readme, /npm install -g @spencer17x\/wtx/);
  assert.match(readme, /command remains `wtx`/);
  assert.match(readmeZh, /npm install -g @spencer17x\/wtx/);
  assert.match(readmeZh, /命令仍然是 `wtx`/);
  assert.match(releasing, /npm package ownership for `@spencer17x\/wtx`/);
  assert.match(releasingZh, /`@spencer17x\/wtx` npm 包名的发布权限/);
});
