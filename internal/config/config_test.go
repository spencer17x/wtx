package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spencer17x/wtx/internal/config"
	"github.com/spencer17x/wtx/internal/core"
)

func TestLoadMergedConfigCombinesUserAndProjectOverrides(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	repoRoot := filepath.Join(root, "repo")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if err := os.WriteFile(filepath.Join(homeDir, ".wtx.json"), []byte(`{
  "worktreeRoot": "/Users/alex/worktrees",
  "strategyOverrides": {
    ".env": "copy",
    ".claude": "symlink"
  },
  "hooks": {
    "beforeCreate": [
      {
        "id": "announce-start",
        "description": "Announce worktree creation",
        "command": "echo",
        "args": ["before-create"]
      }
    ]
  },
  "setupTemplates": {
    "node-pnpm": [
      {
        "id": "node-pnpm-frozen",
        "description": "Install Node.js dependencies with pnpm using the lockfile",
        "command": "pnpm",
        "args": ["install", "--frozen-lockfile"]
      }
    ]
  }
}`), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, ".wtx.json"), []byte(`{
  "strategyOverrides": {
    ".env": "skip",
    "node_modules": "setup"
  },
  "hooks": {
    "afterCreate": [
      {
        "id": "announce-finish",
        "description": "Announce worktree completion",
        "command": "echo",
        "args": ["after-create"]
      }
    ]
  },
  "setupTemplates": {
    "python-uv-sync": [
      {
        "id": "python-uv-sync-locked",
        "description": "Sync Python dependencies with uv using the lockfile",
        "command": "uv",
        "args": ["sync", "--frozen"]
      }
    ]
  }
}`), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	got, err := config.LoadMergedConfig(repoRoot, homeDir)
	if err != nil {
		t.Fatalf("load merged config: %v", err)
	}

	if got.WorktreeRoot != "/Users/alex/worktrees" {
		t.Fatalf("worktree root = %q, want %q", got.WorktreeRoot, "/Users/alex/worktrees")
	}

	wantOverrides := map[string]core.Strategy{
		".env":         core.StrategySkip,
		".claude":      core.StrategySymlink,
		"node_modules": core.StrategySetup,
	}

	if !reflect.DeepEqual(got.StrategyOverrides, wantOverrides) {
		t.Fatalf("strategy overrides = %#v, want %#v", got.StrategyOverrides, wantOverrides)
	}

	wantTemplates := map[string][]core.SetupCommand{
		"node-pnpm": {
			{
				ID:          "node-pnpm-frozen",
				Description: "Install Node.js dependencies with pnpm using the lockfile",
				Command:     "pnpm",
				Args:        []string{"install", "--frozen-lockfile"},
			},
		},
		"python-uv-sync": {
			{
				ID:          "python-uv-sync-locked",
				Description: "Sync Python dependencies with uv using the lockfile",
				Command:     "uv",
				Args:        []string{"sync", "--frozen"},
			},
		},
	}

	if !reflect.DeepEqual(got.SetupTemplates, wantTemplates) {
		t.Fatalf("setup templates = %#v, want %#v", got.SetupTemplates, wantTemplates)
	}

	wantHooks := map[string][]core.CommandHook{
		"beforeCreate": {
			{
				ID:          "announce-start",
				Description: "Announce worktree creation",
				Command:     "echo",
				Args:        []string{"before-create"},
			},
		},
		"afterCreate": {
			{
				ID:          "announce-finish",
				Description: "Announce worktree completion",
				Command:     "echo",
				Args:        []string{"after-create"},
			},
		},
	}

	if !reflect.DeepEqual(got.Hooks, wantHooks) {
		t.Fatalf("hooks = %#v, want %#v", got.Hooks, wantHooks)
	}
}

func TestLoadMergedConfigRejectsInvalidStrategy(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	repoRoot := filepath.Join(root, "repo")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, ".wtx.json"), []byte(`{
  "strategyOverrides": {
    ".env": "share"
  }
}`), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	if _, err := config.LoadMergedConfig(repoRoot, homeDir); err == nil {
		t.Fatal("expected invalid strategy error, got nil")
	}
}

func TestLoadMergedConfigRejectsInvalidSetupTemplate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	repoRoot := filepath.Join(root, "repo")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, ".wtx.json"), []byte(`{
  "setupTemplates": {
    "node-pnpm": [
      {
        "id": "node-pnpm-frozen",
        "description": "broken template",
        "args": ["install"]
      }
    ]
  }
}`), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	if _, err := config.LoadMergedConfig(repoRoot, homeDir); err == nil {
		t.Fatal("expected invalid setup template error, got nil")
	}
}

func TestLoadMergedConfigRejectsInvalidHook(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	repoRoot := filepath.Join(root, "repo")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, ".wtx.json"), []byte(`{
  "hooks": {
    "beforeCreate": [
      {
        "id": "broken-hook",
        "description": "broken hook"
      }
    ]
  }
}`), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	if _, err := config.LoadMergedConfig(repoRoot, homeDir); err == nil {
		t.Fatal("expected invalid hook error, got nil")
	}
}
