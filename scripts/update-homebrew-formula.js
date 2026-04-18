const fs = require("node:fs");
const path = require("node:path");

function renderFormula(template, { version, sha256, url }) {
  return template
    .replaceAll("{{VERSION}}", version)
    .replaceAll("{{SHA256}}", sha256)
    .replaceAll("{{URL}}", url);
}

function validateCliArgs(args) {
  const [templatePath, outputPath, version, sha256, url] = args;
  if (!templatePath || !outputPath || !version || !sha256 || !url) {
    throw new Error(
      "Usage: node scripts/update-homebrew-formula.js <templatePath> <outputPath> <version> <sha256> <url>",
    );
  }
}

function updateFormula({
  templatePath,
  outputPath,
  version,
  sha256,
  url,
}) {
  const template = fs.readFileSync(templatePath, "utf8");
  const formula = renderFormula(template, { version, sha256, url });
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, formula);
}

module.exports = {
  renderFormula,
  validateCliArgs,
  updateFormula,
};

if (require.main === module) {
  try {
    const args = process.argv.slice(2);
    validateCliArgs(args);
    const [templatePath, outputPath, version, sha256, url] = args;
    updateFormula({ templatePath, outputPath, version, sha256, url });
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }
}
