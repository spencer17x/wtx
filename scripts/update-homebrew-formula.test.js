const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const {
  renderFormula,
  validateCliArgs,
  updateFormula,
} = require("./update-homebrew-formula");

test("renders the formula with version, sha256, and url", () => {
  const template = [
    "class Wtx < Formula",
    "  url \"{{URL}}\"",
    "  sha256 \"{{SHA256}}\"",
    "  version \"{{VERSION}}\"",
    "end",
  ].join("\n");

  const formula = renderFormula(template, {
    version: "0.1.0",
    sha256: "abc123",
    url: "https://github.com/spencer17x/wtx/archive/refs/tags/v0.1.0.tar.gz",
  });

  assert.match(formula, /version "0.1.0"/);
  assert.match(formula, /sha256 "abc123"/);
  assert.match(formula, /url "https:\/\/github.com\/spencer17x\/wtx\/archive\/refs\/tags\/v0\.1\.0\.tar\.gz"/);
});

test("rejects missing cli arguments", () => {
  assert.throws(
    () => validateCliArgs(["template", "output", "0.1.0"]),
    /Usage: node scripts\/update-homebrew-formula.js <templatePath> <outputPath> <version> <sha256> <url>/,
  );
});

test("writes the formula to an arbitrary output path", () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-formula-"));
  const templatePath = path.join(tempDir, "wtx.rb.tmpl");
  const outputPath = path.join(tempDir, "Formula", "wtx.rb");

  fs.writeFileSync(
    templatePath,
    [
      "class Wtx < Formula",
      "  url \"{{URL}}\"",
      "  sha256 \"{{SHA256}}\"",
      "  version \"{{VERSION}}\"",
      "end",
    ].join("\n"),
  );

  updateFormula({
    templatePath,
    outputPath,
    version: "0.1.0",
    sha256: "abc123",
    url: "https://github.com/spencer17x/wtx/archive/refs/tags/v0.1.0.tar.gz",
  });

  assert.equal(fs.readFileSync(outputPath, "utf8").includes('version "0.1.0"'), true);
  assert.equal(fs.readFileSync(outputPath, "utf8").includes('sha256 "abc123"'), true);
});
