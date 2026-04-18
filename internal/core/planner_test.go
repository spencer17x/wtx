package core_test

import (
	"reflect"
	"testing"

	"github.com/spencer17x/wtx/internal/core"
)

func TestBuildInitializationPlanAssignsDefaultStrategies(t *testing.T) {
	t.Parallel()

	got := core.BuildInitializationPlan(core.BuildInitializationPlanInput{
		IgnoredPaths: []string{
			".env",
			".vscode/",
			".claude/",
			"node_modules/",
			"dist/",
			"coverage/",
		},
		Mode: core.InitializationModeDefault,
	})

	want := []core.PlanEntry{
		{Path: ".env", Strategy: core.StrategyCopy},
		{Path: ".vscode", Strategy: core.StrategyCopy},
		{Path: ".claude", Strategy: core.StrategySymlink},
		{Path: "node_modules", Strategy: core.StrategySetup},
		{Path: "dist", Strategy: core.StrategySkip},
		{Path: "coverage", Strategy: core.StrategySkip},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %#v, want %#v", got, want)
	}
}

func TestBuildInitializationPlanSupportsCustomOverrides(t *testing.T) {
	t.Parallel()

	got := core.BuildInitializationPlan(core.BuildInitializationPlanInput{
		IgnoredPaths: []string{".env", ".cursor/", "node_modules/"},
		Mode:         core.InitializationModeCustom,
		CustomStrategies: map[string]core.Strategy{
			".env":         core.StrategySkip,
			".cursor":      core.StrategyCopy,
			"node_modules": core.StrategySetup,
		},
	})

	want := []core.PlanEntry{
		{Path: ".env", Strategy: core.StrategySkip},
		{Path: ".cursor", Strategy: core.StrategyCopy},
		{Path: "node_modules", Strategy: core.StrategySetup},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %#v, want %#v", got, want)
	}
}

func TestBuildInitializationPlanSkipsEverythingInNoneMode(t *testing.T) {
	t.Parallel()

	got := core.BuildInitializationPlan(core.BuildInitializationPlanInput{
		IgnoredPaths: []string{".env", ".claude/", "node_modules/"},
		Mode:         core.InitializationModeNone,
	})

	want := []core.PlanEntry{
		{Path: ".env", Strategy: core.StrategySkip},
		{Path: ".claude", Strategy: core.StrategySkip},
		{Path: "node_modules", Strategy: core.StrategySkip},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %#v, want %#v", got, want)
	}
}
