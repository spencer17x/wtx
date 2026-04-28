package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spencer17x/wtx/internal/core"
)

type rawConfig struct {
	WorktreeRoot      string                       `json:"worktreeRoot"`
	StrategyOverrides map[string]string            `json:"strategyOverrides"`
	SetupTemplates    map[string][]rawSetupCommand `json:"setupTemplates"`
	Hooks             map[string][]rawSetupCommand `json:"hooks"`
}

type Config struct {
	WorktreeRoot      string
	StrategyOverrides map[string]core.Strategy
	SetupTemplates    map[string][]core.SetupCommand
	Hooks             map[string][]core.CommandHook
}

type rawSetupCommand struct {
	ID               string   `json:"id"`
	Description      string   `json:"description"`
	Command          string   `json:"command"`
	Args             []string `json:"args"`
	WorkingDirectory string   `json:"cwd"`
}

func normalizePathKey(value string) string {
	normalized := strings.ReplaceAll(value, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimRight(normalized, "/")
	if normalized == "." {
		return ""
	}
	return normalized
}

func parseStrategy(value string) (core.Strategy, error) {
	switch core.Strategy(value) {
	case core.StrategyCopy, core.StrategySymlink, core.StrategySkip, core.StrategySetup:
		return core.Strategy(value), nil
	default:
		return "", fmt.Errorf("invalid strategy %q", value)
	}
}

func mergeConfig(current Config, next Config) Config {
	if next.WorktreeRoot != "" {
		current.WorktreeRoot = next.WorktreeRoot
	}

	if current.StrategyOverrides == nil {
		current.StrategyOverrides = map[string]core.Strategy{}
	}
	if current.SetupTemplates == nil {
		current.SetupTemplates = map[string][]core.SetupCommand{}
	}
	if current.Hooks == nil {
		current.Hooks = map[string][]core.CommandHook{}
	}

	for key, value := range next.StrategyOverrides {
		current.StrategyOverrides[key] = value
	}
	for key, value := range next.SetupTemplates {
		current.SetupTemplates[key] = value
	}
	for key, value := range next.Hooks {
		current.Hooks[key] = value
	}

	return current
}

func parseSetupTemplate(filePath string, key string, commands []rawSetupCommand) ([]core.SetupCommand, error) {
	parsed := make([]core.SetupCommand, 0, len(commands))

	for _, command := range commands {
		if strings.TrimSpace(command.ID) == "" {
			return nil, fmt.Errorf("%s: setup template %q requires a non-empty id", filePath, key)
		}
		if strings.TrimSpace(command.Description) == "" {
			return nil, fmt.Errorf("%s: setup template %q requires a non-empty description", filePath, key)
		}
		if strings.TrimSpace(command.Command) == "" {
			return nil, fmt.Errorf("%s: setup template %q requires a non-empty command", filePath, key)
		}

		parsed = append(parsed, core.SetupCommand{
			ID:               command.ID,
			Description:      command.Description,
			Command:          command.Command,
			Args:             command.Args,
			WorkingDirectory: normalizePathKey(command.WorkingDirectory),
		})
	}

	return parsed, nil
}

func parseHookCommands(filePath string, key string, commands []rawSetupCommand) ([]core.CommandHook, error) {
	parsed := make([]core.CommandHook, 0, len(commands))

	for _, command := range commands {
		if strings.TrimSpace(command.ID) == "" {
			return nil, fmt.Errorf("%s: hook %q requires a non-empty id", filePath, key)
		}
		if strings.TrimSpace(command.Description) == "" {
			return nil, fmt.Errorf("%s: hook %q requires a non-empty description", filePath, key)
		}
		if strings.TrimSpace(command.Command) == "" {
			return nil, fmt.Errorf("%s: hook %q requires a non-empty command", filePath, key)
		}

		parsed = append(parsed, core.CommandHook{
			ID:          command.ID,
			Description: command.Description,
			Command:     command.Command,
			Args:        command.Args,
		})
	}

	return parsed, nil
}

func loadConfigFile(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", filePath, err)
	}

	parsed := Config{
		WorktreeRoot:      raw.WorktreeRoot,
		StrategyOverrides: map[string]core.Strategy{},
		SetupTemplates:    map[string][]core.SetupCommand{},
		Hooks:             map[string][]core.CommandHook{},
	}

	for key, value := range raw.StrategyOverrides {
		strategy, err := parseStrategy(value)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", filePath, err)
		}

		normalizedKey := normalizePathKey(key)
		if normalizedKey == "" {
			continue
		}

		parsed.StrategyOverrides[normalizedKey] = strategy
	}

	for key, commands := range raw.SetupTemplates {
		parsedCommands, err := parseSetupTemplate(filePath, key, commands)
		if err != nil {
			return Config{}, err
		}

		parsed.SetupTemplates[key] = parsedCommands
	}

	for key, commands := range raw.Hooks {
		parsedCommands, err := parseHookCommands(filePath, key, commands)
		if err != nil {
			return Config{}, err
		}

		parsed.Hooks[key] = parsedCommands
	}

	return parsed, nil
}

func LoadMergedConfig(repoRoot string, homeDir string) (Config, error) {
	merged := Config{
		StrategyOverrides: map[string]core.Strategy{},
	}

	userConfig, err := loadConfigFile(filepath.Join(homeDir, ".wtx.json"))
	if err != nil {
		return Config{}, err
	}
	merged = mergeConfig(merged, userConfig)

	projectConfig, err := loadConfigFile(filepath.Join(repoRoot, ".wtx.json"))
	if err != nil {
		return Config{}, err
	}
	merged = mergeConfig(merged, projectConfig)

	return merged, nil
}
