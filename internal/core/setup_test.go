package core_test

import (
	"reflect"
	"testing"

	"github.com/spencer17x/wtx/internal/core"
)

func TestDetectSetupCommandsPrefersPNPM(t *testing.T) {
	t.Parallel()

	got := core.DetectSetupCommands([]string{"package.json", "pnpm-lock.yaml"})
	want := []core.SetupCommand{
		{
			ID:          "node-pnpm",
			Description: "Install Node.js dependencies with pnpm",
			Command:     "pnpm",
			Args:        []string{"install"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestDetectSetupCommandsPrefersBunWhenBunLockIsPresent(t *testing.T) {
	t.Parallel()

	got := core.DetectSetupCommands([]string{"package.json", "bun.lockb", "pnpm-lock.yaml"})
	want := []core.SetupCommand{
		{
			ID:          "node-bun",
			Description: "Install Node.js dependencies with Bun",
			Command:     "bun",
			Args:        []string{"install"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestDetectSetupCommandsReturnsPythonVirtualenvFlow(t *testing.T) {
	t.Parallel()

	got := core.DetectSetupCommands([]string{"pyproject.toml", "requirements.txt"})
	want := []core.SetupCommand{
		{
			ID:          "python-venv",
			Description: "Create a Python virtual environment",
			Command:     "python",
			Args:        []string{"-m", "venv", ".venv"},
		},
		{
			ID:          "python-install-requirements",
			Description: "Install Python dependencies from requirements.txt",
			Command:     ".venv/bin/python",
			Args:        []string{"-m", "pip", "install", "-r", "requirements.txt"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestDetectSetupCommandsPrefersUVAndPoetryOverGenericPythonFlow(t *testing.T) {
	t.Parallel()

	uvCommands := core.DetectSetupCommands([]string{"pyproject.toml", "uv.lock"})
	uvWant := []core.SetupCommand{
		{
			ID:          "python-uv-sync",
			Description: "Sync Python dependencies with uv",
			Command:     "uv",
			Args:        []string{"sync"},
		},
	}
	if !reflect.DeepEqual(uvCommands, uvWant) {
		t.Fatalf("uv commands = %#v, want %#v", uvCommands, uvWant)
	}

	poetryCommands := core.DetectSetupCommands([]string{"pyproject.toml", "poetry.lock"})
	poetryWant := []core.SetupCommand{
		{
			ID:          "python-poetry-install",
			Description: "Install Python dependencies with Poetry",
			Command:     "poetry",
			Args:        []string{"install"},
		},
	}
	if !reflect.DeepEqual(poetryCommands, poetryWant) {
		t.Fatalf("poetry commands = %#v, want %#v", poetryCommands, poetryWant)
	}
}

func TestDetectSetupCommandsPrefersPipenvWhenPipfileIsPresent(t *testing.T) {
	t.Parallel()

	got := core.DetectSetupCommands([]string{"Pipfile", "Pipfile.lock"})
	want := []core.SetupCommand{
		{
			ID:          "python-pipenv-install",
			Description: "Install Python dependencies with Pipenv",
			Command:     "pipenv",
			Args:        []string{"install"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestDetectSetupCommandsFindsGoRustMavenAndGradle(t *testing.T) {
	t.Parallel()

	got := core.DetectSetupCommands([]string{"go.mod", "Cargo.toml", "pom.xml", "build.gradle"})
	want := []core.SetupCommand{
		{
			ID:          "go-download",
			Description: "Download Go module dependencies",
			Command:     "go",
			Args:        []string{"mod", "download"},
		},
		{
			ID:          "rust-fetch",
			Description: "Fetch Rust dependencies",
			Command:     "cargo",
			Args:        []string{"fetch"},
		},
		{
			ID:          "java-maven-resolve",
			Description: "Resolve Maven dependencies",
			Command:     "mvn",
			Args:        []string{"dependency:resolve"},
		},
		{
			ID:          "java-gradle-build",
			Description: "Build Gradle dependencies",
			Command:     "gradle",
			Args:        []string{"build"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestDetectSetupCommandsPrefersJavaWrappersWhenPresent(t *testing.T) {
	t.Parallel()

	got := core.DetectSetupCommands([]string{"pom.xml", "mvnw", "build.gradle", "gradlew"})
	want := []core.SetupCommand{
		{
			ID:          "java-maven-wrapper-resolve",
			Description: "Resolve Maven dependencies with the Maven wrapper",
			Command:     "./mvnw",
			Args:        []string{"dependency:resolve"},
		},
		{
			ID:          "java-gradle-wrapper-build",
			Description: "Build Gradle dependencies with the Gradle wrapper",
			Command:     "./gradlew",
			Args:        []string{"build"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestApplySetupTemplatesReplacesDetectedCommandsByID(t *testing.T) {
	t.Parallel()

	got := core.ApplySetupTemplates(
		[]core.SetupCommand{
			{
				ID:          "node-pnpm",
				Description: "Install Node.js dependencies with pnpm",
				Command:     "pnpm",
				Args:        []string{"install"},
			},
			{
				ID:          "go-download",
				Description: "Download Go module dependencies",
				Command:     "go",
				Args:        []string{"mod", "download"},
			},
		},
		map[string][]core.SetupCommand{
			"node-pnpm": {
				{
					ID:          "node-pnpm-frozen",
					Description: "Install Node.js dependencies with pnpm using the lockfile",
					Command:     "pnpm",
					Args:        []string{"install", "--frozen-lockfile"},
				},
			},
		},
	)

	want := []core.SetupCommand{
		{
			ID:          "node-pnpm-frozen",
			Description: "Install Node.js dependencies with pnpm using the lockfile",
			Command:     "pnpm",
			Args:        []string{"install", "--frozen-lockfile"},
		},
		{
			ID:          "go-download",
			Description: "Download Go module dependencies",
			Command:     "go",
			Args:        []string{"mod", "download"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}
