#!/usr/bin/env node

const fs = require("node:fs");
const { spawnSync } = require("node:child_process");
const path = require("node:path");

const { resolveBinaryPath } = require("../install");

const binaryPath = resolveBinaryPath(path.join(__dirname, "..", ".."), "wtx");

if (!fs.existsSync(binaryPath)) {
  console.error(`Missing installed native binary at ${binaryPath}`);
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
});

if (result.error) {
  throw result.error;
}

if (result.signal) {
  process.kill(process.pid, result.signal);
}

process.exit(result.status ?? 0);
