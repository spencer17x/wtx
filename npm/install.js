const fs = require("node:fs");
const path = require("node:path");
const https = require("node:https");
const crypto = require("node:crypto");
const { execFileSync } = require("node:child_process");

const { getAssetInfo } = require("./platform");

const CHECKSUM_ASSET_NAME = "checksums.txt";

function buildReleaseAssetUrl({ owner, repo, version, archive }) {
  return `https://github.com/${owner}/${repo}/releases/download/v${version}/${archive}`;
}

function buildChecksumsAssetUrl({ owner, repo, version }) {
  return buildReleaseAssetUrl({
    owner,
    repo,
    version,
    archive: CHECKSUM_ASSET_NAME,
  });
}

function resolveBinaryPath(packageRoot, binaryName) {
  return path.join(packageRoot, "npm", "bin", binaryName);
}

function shouldSkipInstall(version) {
  return version === "0.0.0-development";
}

function cleanupDestination(destination, file, rmImpl) {
  return new Promise((resolve) => {
    const finalize = () => rmImpl(destination, { force: true }, () => resolve());
    if (file && typeof file.close === "function") {
      file.close(finalize);
      return;
    }
    finalize();
  });
}

function downloadFile(
  url,
  destination,
  {
    getImpl = https.get,
    createWriteStreamImpl = fs.createWriteStream,
    rmImpl = fs.rm,
    redirectCount = 0,
    timeoutMs = 30000,
  } = {},
) {
  return new Promise((resolve, reject) => {
    const file = createWriteStreamImpl(destination);
    let settled = false;

    const fail = (error) => {
      if (settled) {
        return;
      }
      settled = true;
      cleanupDestination(destination, file, rmImpl).then(() => reject(error));
    };

    const succeed = () => {
      if (settled) {
        return;
      }
      settled = true;
      if (file && typeof file.close === "function") {
        file.close(() => resolve());
        return;
      }
      resolve();
    };

    const request = getImpl(url, (response) => {
      if (settled) {
        if (response && typeof response.resume === "function") {
          response.resume();
        }
        return;
      }

      const isRedirect = response.statusCode >= 300 && response.statusCode < 400;
      if (isRedirect && response.headers.location) {
        settled = true;

        (async () => {
          await cleanupDestination(destination, file, rmImpl);
          if (redirectCount >= 5) {
            throw new Error(`Too many redirects while downloading ${url}`);
          }

          return downloadFile(
            new URL(response.headers.location, url).toString(),
            destination,
            {
              getImpl,
              createWriteStreamImpl,
              rmImpl,
              redirectCount: redirectCount + 1,
              timeoutMs,
            },
          );
        })().then(resolve, reject);
        return;
      }

      if (response.statusCode !== 200) {
        fail(new Error(`Failed to download ${url}: ${response.statusCode}`));
        return;
      }

      response.on("error", fail);
      if (file && typeof file.on === "function") {
        file.on("error", fail);
      }

      response.pipe(file);
      if (file && typeof file.on === "function") {
        file.on("finish", succeed);
      } else {
        succeed();
      }
    });

    if (request && typeof request.on === "function") {
      request.on("error", fail);
    }
    if (request && typeof request.setTimeout === "function") {
      request.setTimeout(timeoutMs, () => {
        const error = new Error(`Timed out downloading ${url}`);
        fail(error);
        if (typeof request.destroy === "function") {
          request.destroy(error);
        }
      });
    }
  });
}

function formatExtractionError(error) {
  if (error && error.code === "ENOENT") {
    return new Error("Unable to extract wtx archive: tar is not available on PATH");
  }

  return new Error(`Unable to extract wtx archive: ${error && error.message ? error.message : "tar failed"}`);
}

function extractArchive(archivePath, outputDir, { execFileSyncImpl = execFileSync } = {}) {
  try {
    execFileSyncImpl("tar", ["-xzf", archivePath, "-C", outputDir]);
  } catch (error) {
    throw formatExtractionError(error);
  }
}

function findArchiveChecksum(checksums, archiveName) {
  for (const line of checksums.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (trimmed === "") {
      continue;
    }

    const [hash, fileName] = trimmed.split(/\s+/);
    if (fileName === archiveName) {
      return hash;
    }
  }

  return "";
}

function sha256File(filePath) {
  return crypto.createHash("sha256").update(fs.readFileSync(filePath)).digest("hex");
}

function verifyArchiveChecksum(archivePath, checksumsPath, archiveName) {
  const expectedHash = findArchiveChecksum(fs.readFileSync(checksumsPath, "utf8"), archiveName);
  if (!expectedHash) {
    throw new Error(`Missing checksum for ${archiveName}`);
  }

  const actualHash = sha256File(archivePath);
  if (actualHash !== expectedHash) {
    throw new Error(`Checksum mismatch for ${archiveName}`);
  }
}

async function installBinary({
  packageRoot = path.resolve(__dirname, ".."),
  version,
  owner = "spencer17x",
  repo = "wtx",
  platform = process.platform,
  arch = process.arch,
  downloadFileImpl = downloadFile,
  execFileSyncImpl = execFileSync,
} = {}) {
  if (shouldSkipInstall(version)) {
    return { skipped: true };
  }

  const asset = getAssetInfo(platform, arch);
  const archivePath = path.join(packageRoot, asset.archive);
  const checksumsPath = path.join(packageRoot, CHECKSUM_ASSET_NAME);
  const binaryPath = resolveBinaryPath(packageRoot, asset.binaryName);

  fs.mkdirSync(path.dirname(binaryPath), { recursive: true });

  const url = buildReleaseAssetUrl({
    owner,
    repo,
    version,
    archive: asset.archive,
  });

  try {
    await downloadFileImpl(url, archivePath);
    await downloadFileImpl(buildChecksumsAssetUrl({ owner, repo, version }), checksumsPath);
    verifyArchiveChecksum(archivePath, checksumsPath, asset.archive);
    extractArchive(archivePath, path.dirname(binaryPath), { execFileSyncImpl });
    fs.chmodSync(binaryPath, 0o755);
  } catch (error) {
    fs.rmSync(binaryPath, { force: true });
    throw error;
  } finally {
    fs.rmSync(archivePath, { force: true });
    fs.rmSync(checksumsPath, { force: true });
  }

  return { skipped: false };
}

module.exports = {
  buildReleaseAssetUrl,
  buildChecksumsAssetUrl,
  downloadFile,
  resolveBinaryPath,
  shouldSkipInstall,
  installBinary,
  extractArchive,
  formatExtractionError,
  verifyArchiveChecksum,
};

if (require.main === module) {
  const version = require("../package.json").version;
  if (shouldSkipInstall(version)) {
    console.log(`Skipping npm install for development version ${version}`);
    process.exit(0);
  }

  installBinary({ version }).catch((error) => {
    console.error(error.message);
    process.exit(1);
  });
}
