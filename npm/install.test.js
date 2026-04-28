const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const crypto = require("node:crypto");
const { PassThrough, Readable, Writable } = require("node:stream");

const {
  buildReleaseAssetUrl,
  downloadFile,
  resolveBinaryPath,
  installBinary,
  shouldSkipInstall,
} = require("./install");

function writeArchiveOrChecksum(url, destination, archiveName, archivePayload) {
  if (url.endsWith("/checksums.txt")) {
    const archiveHash = crypto
      .createHash("sha256")
      .update(archivePayload)
      .digest("hex");
    fs.writeFileSync(destination, `${archiveHash}  ${archiveName}\n`);
    return;
  }

  fs.writeFileSync(destination, archivePayload);
}

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

test("times out stalled downloads and destroys the request", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const destination = path.join(tempDir, "asset.tgz");
  let destroyed = false;

  await assert.rejects(
    downloadFile("https://example.com/archive.tgz", destination, {
      timeoutMs: 1,
      getImpl: (_url, callback) => {
        const response = Readable.from(["payload"]);
        response.statusCode = 200;
        response.headers = {};
        process.nextTick(() => callback(response));
        return {
          on() {},
          setTimeout(ms, onTimeout) {
            assert.equal(ms, 1);
            onTimeout();
          },
          destroy(error) {
            destroyed = true;
            assert.match(error.message, /Timed out downloading/);
          },
        };
      },
      createWriteStreamImpl: () => {
        const file = new PassThrough();
        file.close = (cb) => process.nextTick(cb);
        return file;
      },
      rmImpl: (_target, _options, cb) => process.nextTick(cb),
    }),
    /Timed out downloading/,
  );

  assert.equal(destroyed, true);
});

test("follows redirects when downloading release assets", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const destination = path.join(tempDir, "asset.tgz");
  const urls = [];
  let createStreamCalls = 0;
  let cleanupCount = 0;

  const getImpl = (url, callback) => {
    urls.push(url);

    if (urls.length === 1) {
      const response = new Readable({ read() {} });
      response.statusCode = 302;
      response.headers = { location: "/download/real-asset.tgz" };
      process.nextTick(() => callback(response));
      return { on() {} };
    }

    const response = Readable.from(["payload"]);
    response.statusCode = 200;
    response.headers = {};
    process.nextTick(() => callback(response));
    return { on() {} };
  };

  const rmImpl = (target, options, cb) => {
    cleanupCount += 1;
    setTimeout(() => {
      fs.rmSync(target, { force: true });
      cb();
    }, 25);
  };

  await downloadFile(
    "https://github.com/spencer17x/wtx/releases/download/v0.1.0/wtx_Darwin_arm64.tar.gz",
    destination,
    {
      getImpl,
      createWriteStreamImpl: (target) => {
        createStreamCalls += 1;
        return fs.createWriteStream(target);
      },
      rmImpl,
    },
  );

  await new Promise((resolve) => setTimeout(resolve, 60));

  assert.deepEqual(urls, [
    "https://github.com/spencer17x/wtx/releases/download/v0.1.0/wtx_Darwin_arm64.tar.gz",
    "https://github.com/download/real-asset.tgz",
  ]);
  assert.equal(createStreamCalls, 2);
  assert.equal(cleanupCount, 1);
  assert.equal(fs.readFileSync(destination, "utf8"), "payload");
  assert.equal(fs.existsSync(destination), true);
});

test("cleans up partial downloads when the response stream fails", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const destination = path.join(tempDir, "asset.tgz");
  const removed = [];
  const response = new PassThrough();
  response.statusCode = 200;
  response.headers = {};
  response.close = (cb) => process.nextTick(cb);

  const getImpl = (url, callback) => {
    process.nextTick(() => callback(response));
    process.nextTick(() => {
      response.write("partial");
      response.destroy(new Error("socket closed"));
    });
    return { on() {} };
  };

  await assert.rejects(
    downloadFile("https://example.com/archive.tgz", destination, {
      getImpl,
      createWriteStreamImpl: () => new PassThrough(),
      rmImpl: (target, options, cb) => {
        removed.push(target);
        process.nextTick(cb);
      },
    }),
    /socket closed/,
  );

  assert.deepEqual(removed, [destination]);
});

test("cleans up partial downloads when the file stream fails", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const destination = path.join(tempDir, "asset.tgz");
  const removed = [];

  const file = new Writable({
    write(_chunk, _encoding, callback) {
      callback(new Error("disk full"));
    },
  });
  file.close = (cb) => process.nextTick(cb);

  const getImpl = (url, callback) => {
    const response = new PassThrough();
    response.statusCode = 200;
    response.headers = {};
    process.nextTick(() => callback(response));
    process.nextTick(() => {
      response.end("partial");
    });
    return { on() {} };
  };

  await assert.rejects(
    downloadFile("https://example.com/archive.tgz", destination, {
      getImpl,
      createWriteStreamImpl: () => file,
      rmImpl: (target, options, cb) => {
        removed.push(target);
        process.nextTick(cb);
      },
    }),
    /disk full/,
  );

  assert.deepEqual(removed, [destination]);
});

test("skips installing in development checkouts", async () => {
  let attempted = false;

  const result = await installBinary({
    version: "0.0.0-development",
    downloadFileImpl: async () => {
      attempted = true;
      throw new Error("should not be called");
    },
  });

  assert.equal(result.skipped, true);
  assert.equal(attempted, false);
  assert.equal(shouldSkipInstall("0.0.0-development"), true);
});

test("verifies downloaded release archives against checksums", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const binaryPath = path.join(tempDir, "npm", "bin", "wtx");
  const archiveName = "wtx_Darwin_arm64.tar.gz";
  const archivePayload = "archive";
  const archiveHash = crypto
    .createHash("sha256")
    .update(archivePayload)
    .digest("hex");
  const downloaded = [];

  const result = await installBinary({
    packageRoot: tempDir,
    version: "0.1.0",
    platform: "darwin",
    arch: "arm64",
    downloadFileImpl: async (url, destination) => {
      downloaded.push(path.basename(url));
      if (url.endsWith("/checksums.txt")) {
        fs.writeFileSync(destination, `${archiveHash}  ${archiveName}\n`);
        return;
      }
      fs.writeFileSync(destination, archivePayload);
    },
    execFileSyncImpl: () => {
      fs.mkdirSync(path.dirname(binaryPath), { recursive: true });
      fs.writeFileSync(binaryPath, "binary");
    },
  });

  assert.equal(result.skipped, false);
  assert.deepEqual(downloaded, [archiveName, "checksums.txt"]);
  assert.equal(fs.existsSync(binaryPath), true);
});

test("rejects release archives with mismatched checksums", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const archiveName = "wtx_Darwin_arm64.tar.gz";
  let extracted = false;

  await assert.rejects(
    installBinary({
      packageRoot: tempDir,
      version: "0.1.0",
      platform: "darwin",
      arch: "arm64",
      downloadFileImpl: async (url, destination) => {
        if (url.endsWith("/checksums.txt")) {
          fs.writeFileSync(destination, `${"0".repeat(64)}  ${archiveName}\n`);
          return;
        }
        fs.writeFileSync(destination, "archive");
      },
      execFileSyncImpl: () => {
        extracted = true;
      },
    }),
    /Checksum mismatch/,
  );

  assert.equal(extracted, false);
});

test("surfaces a clear error when tar is unavailable", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const archivePath = path.join(tempDir, "wtx_Darwin_arm64.tar.gz");
  const archivePayload = "archive";

  await assert.rejects(
    installBinary({
      packageRoot: tempDir,
      version: "0.1.0",
      platform: "darwin",
      arch: "arm64",
      downloadFileImpl: async (url, destination) => {
        writeArchiveOrChecksum(url, destination, "wtx_Darwin_arm64.tar.gz", archivePayload);
      },
      execFileSyncImpl: () => {
        const error = new Error("spawn tar ENOENT");
        error.code = "ENOENT";
        throw error;
      },
    }),
    /Unable to extract wtx archive: tar is not available on PATH/,
  );

  assert.equal(fs.existsSync(archivePath), false);
});

test("surfaces a clear error when tar extraction fails", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const archivePayload = "archive";

  await assert.rejects(
    installBinary({
      packageRoot: tempDir,
      version: "0.1.0",
      platform: "darwin",
      arch: "arm64",
      downloadFileImpl: async (url, destination) => {
        writeArchiveOrChecksum(url, destination, "wtx_Darwin_arm64.tar.gz", archivePayload);
      },
      execFileSyncImpl: () => {
        throw new Error("tar exited 2");
      },
    }),
    /Unable to extract wtx archive: tar exited 2/,
  );
});

test("removes the binary when extraction fails after writing it", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const binaryPath = path.join(tempDir, "npm", "bin", "wtx");
  const archivePayload = "archive";

  await assert.rejects(
    installBinary({
      packageRoot: tempDir,
      version: "0.1.0",
      platform: "darwin",
      arch: "arm64",
      downloadFileImpl: async (url, destination) => {
        fs.mkdirSync(path.dirname(destination), { recursive: true });
        writeArchiveOrChecksum(url, destination, "wtx_Darwin_arm64.tar.gz", archivePayload);
      },
      execFileSyncImpl: () => {
        fs.mkdirSync(path.dirname(binaryPath), { recursive: true });
        fs.writeFileSync(binaryPath, "partial binary");
        throw new Error("tar exited 2");
      },
    }),
    /Unable to extract wtx archive: tar exited 2/,
  );

  assert.equal(fs.existsSync(binaryPath), false);
});
