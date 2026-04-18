const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { PassThrough, Readable, Writable } = require("node:stream");

const {
  buildReleaseAssetUrl,
  downloadFile,
  resolveBinaryPath,
  installBinary,
  shouldSkipInstall,
} = require("./install");

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

test("follows redirects when downloading release assets", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const destination = path.join(tempDir, "asset.tgz");
  const urls = [];

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

  await downloadFile(
    "https://github.com/spencer17x/wtx/releases/download/v0.1.0/wtx_Darwin_arm64.tar.gz",
    destination,
    { getImpl },
  );

  assert.deepEqual(urls, [
    "https://github.com/spencer17x/wtx/releases/download/v0.1.0/wtx_Darwin_arm64.tar.gz",
    "https://github.com/download/real-asset.tgz",
  ]);
  assert.equal(fs.readFileSync(destination, "utf8"), "payload");
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

test("surfaces a clear error when tar is unavailable", async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "wtx-install-"));
  const archivePath = path.join(tempDir, "wtx_Darwin_arm64.tar.gz");

  await assert.rejects(
    installBinary({
      packageRoot: tempDir,
      version: "0.1.0",
      platform: "darwin",
      arch: "arm64",
      downloadFileImpl: async (_url, destination) => {
        fs.writeFileSync(destination, "archive");
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

  await assert.rejects(
    installBinary({
      packageRoot: tempDir,
      version: "0.1.0",
      platform: "darwin",
      arch: "arm64",
      downloadFileImpl: async (_url, destination) => {
        fs.writeFileSync(destination, "archive");
      },
      execFileSyncImpl: () => {
        throw new Error("tar exited 2");
      },
    }),
    /Unable to extract wtx archive: tar exited 2/,
  );
});
