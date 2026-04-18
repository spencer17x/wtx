const test = require("node:test");
const assert = require("node:assert/strict");

const {
  renderFormula,
} = require("./update-homebrew-formula");

test("renders the formula with version and sha256", () => {
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
});
