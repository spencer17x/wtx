package core

import (
	"path"
	"strings"
)

type BuildInitializationPlanInput struct {
	IgnoredPaths     []string
	Mode             InitializationMode
	CustomStrategies map[string]Strategy
}

var copyNames = map[string]struct{}{
	".vscode":         {},
	".idea":           {},
	".env":            {},
	".python-version": {},
	".tool-versions":  {},
}

var symlinkNames = map[string]struct{}{
	".claude": {},
	".cursor": {},
}

var setupNames = map[string]struct{}{
	"node_modules": {},
	".venv":        {},
	"venv":         {},
	".gradle":      {},
	"target":       {},
}

var skipNames = map[string]struct{}{
	"dist":          {},
	".next":         {},
	"build":         {},
	"coverage":      {},
	"out":           {},
	".cache":        {},
	".parcel-cache": {},
	".svelte-kit":   {},
	"logs":          {},
	"log":           {},
	"tmp":           {},
	"temp":          {},
}

func normalizeCandidatePath(value string) string {
	normalized := strings.ReplaceAll(value, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimRight(normalized, "/")
	return normalized
}

func isConfigFile(name string) bool {
	if _, ok := copyNames[name]; ok {
		return true
	}

	return strings.HasPrefix(name, ".env.") ||
		strings.HasSuffix(name, ".local") ||
		strings.HasSuffix(name, ".pem") ||
		strings.HasSuffix(name, ".key") ||
		strings.HasSuffix(name, ".crt")
}

func ClassifyIgnoredPath(candidatePath string) Strategy {
	normalizedPath := normalizeCandidatePath(candidatePath)
	baseName := path.Base(normalizedPath)

	if _, ok := setupNames[baseName]; ok {
		return StrategySetup
	}

	if _, ok := symlinkNames[baseName]; ok {
		return StrategySymlink
	}

	if _, ok := skipNames[baseName]; ok || strings.HasSuffix(baseName, ".log") || strings.HasSuffix(baseName, ".tmp") {
		return StrategySkip
	}

	if isConfigFile(baseName) {
		return StrategyCopy
	}

	if strings.HasPrefix(baseName, ".") {
		return StrategyCopy
	}

	return StrategyCopy
}

func BuildInitializationPlan(input BuildInitializationPlanInput) []PlanEntry {
	overrides := make(map[string]Strategy, len(input.CustomStrategies))
	for key, value := range input.CustomStrategies {
		overrides[normalizeCandidatePath(key)] = value
	}

	seen := map[string]struct{}{}
	plan := make([]PlanEntry, 0, len(input.IgnoredPaths))

	for _, rawPath := range input.IgnoredPaths {
		normalizedPath := normalizeCandidatePath(rawPath)
		if normalizedPath == "" {
			continue
		}

		if _, ok := seen[normalizedPath]; ok {
			continue
		}
		seen[normalizedPath] = struct{}{}

		strategy := StrategySkip
		if input.Mode != InitializationModeNone {
			if override, ok := overrides[normalizedPath]; ok {
				strategy = override
			} else {
				strategy = ClassifyIgnoredPath(normalizedPath)
			}
		}

		plan = append(plan, PlanEntry{
			Path:     normalizedPath,
			Strategy: strategy,
		})
	}

	return plan
}
