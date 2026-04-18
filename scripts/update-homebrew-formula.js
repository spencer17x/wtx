const fs = require("node:fs");
const path = require("node:path");

function renderFormula(template, { version, sha256, url }) {
  return template
    .replaceAll("{{VERSION}}", version)
    .replaceAll("{{SHA256}}", sha256)
    .replaceAll("{{URL}}", url);
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
  updateFormula,
};

if (require.main === module) {
  const [templatePath, outputPath, version, sha256, url] = process.argv.slice(2);
  updateFormula({ templatePath, outputPath, version, sha256, url });
}
