#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const binaryPath = path.join(__dirname, "wtx");
const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
});

if (result.error) {
  throw result.error;
}

process.exit(result.status ?? 0);
