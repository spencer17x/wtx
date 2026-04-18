const fs = require("node:fs");
const path = require("node:path");
const https = require("node:https");
const { execFileSync } = require("node:child_process");

const { getAssetInfo } = require("./platform");

function buildReleaseAssetUrl({ owner, repo, version, archive }) {
  return `https://github.com/${owner}/${repo}/releases/download/v${version}/${archive}`;
}

function resolveBinaryPath(packageRoot, binaryName) {
  return path.join(packageRoot, "npm", "bin", binaryName);
}

function downloadFile(url, destination) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destination);

    https.get(url, (response) => {
      if (response.statusCode !== 200) {
        file.close(() => {
          fs.rm(destination, { force: true }, () => {});
        });
        reject(new Error(`Failed to download ${url}: ${response.statusCode}`));
        return;
      }

      response.pipe(file);
      file.on("finish", () => {
        file.close(resolve);
      });
    }).on("error", (error) => {
      file.close(() => {
        fs.rm(destination, { force: true }, () => {});
      });
      reject(error);
    });
  });
}

async function installBinary({
  packageRoot = path.resolve(__dirname, ".."),
  version,
  owner = "spencer17x",
  repo = "wtx",
  platform = process.platform,
  arch = process.arch,
} = {}) {
  const asset = getAssetInfo(platform, arch);
  const archivePath = path.join(packageRoot, asset.archive);
  const binaryPath = resolveBinaryPath(packageRoot, asset.binaryName);

  fs.mkdirSync(path.dirname(binaryPath), { recursive: true });

  const url = buildReleaseAssetUrl({
    owner,
    repo,
    version,
    archive: asset.archive,
  });

  await downloadFile(url, archivePath);
  execFileSync("tar", ["-xzf", archivePath, "-C", path.dirname(binaryPath)]);
  fs.chmodSync(binaryPath, 0o755);
  fs.rmSync(archivePath, { force: true });
}

module.exports = {
  buildReleaseAssetUrl,
  resolveBinaryPath,
  installBinary,
};

if (require.main === module) {
  const version = require("../package.json").version;
  installBinary({ version }).catch((error) => {
    console.error(error.message);
    process.exit(1);
  });
}
