const SUPPORTED_PLATFORMS = {
  darwin: {
    arm64: {
      goos: "Darwin",
      goarch: "arm64",
    },
    x64: {
      goos: "Darwin",
      goarch: "x86_64",
    },
  },
  linux: {
    arm64: {
      goos: "Linux",
      goarch: "arm64",
    },
    x64: {
      goos: "Linux",
      goarch: "x86_64",
    },
  },
};

function getAssetInfo(platform, arch) {
  const platformEntry = SUPPORTED_PLATFORMS[platform];
  const archEntry = platformEntry && platformEntry[arch];

  if (!archEntry) {
    throw new Error(`Unsupported platform: ${platform} ${arch}`);
  }

  return {
    ...archEntry,
    archive: `wtx_${archEntry.goos}_${archEntry.goarch}.tar.gz`,
    binaryName: "wtx",
  };
}

function getReleaseArchives() {
  const archives = [];

  for (const [platform, architectures] of Object.entries(SUPPORTED_PLATFORMS)) {
    for (const arch of Object.keys(architectures)) {
      archives.push(getAssetInfo(platform, arch).archive);
    }
  }

  return archives;
}

function getVersionFromTag(tag) {
  const match = /^v((?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*))$/.exec(tag);
  if (!match) {
    throw new Error("Expected a tag in the form vX.Y.Z");
  }

  return match[1];
}

module.exports = {
  getAssetInfo,
  getReleaseArchives,
  getVersionFromTag,
};
