const fs = require("node:fs");
const { getReleaseArchives } = require("../npm/platform");

const CHECKSUM_ASSET_NAME = "checksums.txt";

function getRequiredReleaseAssetNames() {
  return [...getReleaseArchives(), CHECKSUM_ASSET_NAME];
}

function getReleaseAssetProblems(release) {
  const assets = Array.isArray(release && release.assets) ? release.assets : [];
  const assetsByName = new Map(assets.map((asset) => [asset.name, asset]));
  const missing = [];
  const empty = [];

  for (const name of getRequiredReleaseAssetNames()) {
    const asset = assetsByName.get(name);
    if (!asset) {
      missing.push(name);
      continue;
    }

    if (!Number.isInteger(asset.size) || asset.size <= 0) {
      empty.push(name);
    }
  }

  return { missing, empty };
}

function assertReleaseAssetsComplete(release) {
  const { missing, empty } = getReleaseAssetProblems(release);
  const problems = [];

  if (missing.length > 0) {
    problems.push(`Missing: ${missing.join(", ")}`);
  }

  if (empty.length > 0) {
    problems.push(`Empty: ${empty.join(", ")}`);
  }

  if (problems.length > 0) {
    throw new Error(`Release assets are incomplete. ${problems.join(". ")}`);
  }
}

function validateCliArgs(args) {
  const [releaseJsonPath] = args;
  if (!releaseJsonPath) {
    throw new Error(
      "Usage: node scripts/check-release-assets.js <releaseJsonPath>",
    );
  }
}

function loadRelease(releaseJsonPath) {
  return JSON.parse(fs.readFileSync(releaseJsonPath, "utf8"));
}

module.exports = {
  assertReleaseAssetsComplete,
  getReleaseAssetProblems,
  getRequiredReleaseAssetNames,
  loadRelease,
  validateCliArgs,
};

if (require.main === module) {
  try {
    const args = process.argv.slice(2);
    validateCliArgs(args);
    const [releaseJsonPath] = args;
    assertReleaseAssetsComplete(loadRelease(releaseJsonPath));
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }
}
