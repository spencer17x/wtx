package core

import (
	"path"
	"strings"
)

func createSetupCommand(id string, description string, command string, args ...string) SetupCommand {
	return SetupCommand{
		ID:          id,
		Description: description,
		Command:     command,
		Args:        args,
	}
}

func normalizeFileSet(files []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(files))
	for _, file := range files {
		normalized[strings.ReplaceAll(file, "\\", "/")] = struct{}{}
	}
	return normalized
}

func hasFile(files map[string]struct{}, expected string) bool {
	for file := range files {
		if file == expected || path.Base(file) == expected {
			return true
		}
	}
	return false
}

func DetectSetupCommands(files []string) []SetupCommand {
	normalizedFiles := normalizeFileSet(files)
	commands := []SetupCommand{}

	if hasFile(normalizedFiles, "package.json") {
		switch {
		case hasFile(normalizedFiles, "bun.lockb") || hasFile(normalizedFiles, "bun.lock"):
			commands = append(commands, createSetupCommand(
				"node-bun",
				"Install Node.js dependencies with Bun",
				"bun",
				"install",
			))
		case hasFile(normalizedFiles, "pnpm-lock.yaml"):
			commands = append(commands, createSetupCommand(
				"node-pnpm",
				"Install Node.js dependencies with pnpm",
				"pnpm",
				"install",
			))
		case hasFile(normalizedFiles, "yarn.lock"):
			commands = append(commands, createSetupCommand(
				"node-yarn",
				"Install Node.js dependencies with yarn",
				"yarn",
				"install",
			))
		default:
			commands = append(commands, createSetupCommand(
				"node-npm",
				"Install Node.js dependencies with npm",
				"npm",
				"install",
			))
		}
	}

	if hasFile(normalizedFiles, "pyproject.toml") || hasFile(normalizedFiles, "requirements.txt") || hasFile(normalizedFiles, "Pipfile") {
		switch {
		case hasFile(normalizedFiles, "uv.lock"):
			commands = append(commands, createSetupCommand(
				"python-uv-sync",
				"Sync Python dependencies with uv",
				"uv",
				"sync",
			))
		case hasFile(normalizedFiles, "poetry.lock"):
			commands = append(commands, createSetupCommand(
				"python-poetry-install",
				"Install Python dependencies with Poetry",
				"poetry",
				"install",
			))
		case hasFile(normalizedFiles, "Pipfile") || hasFile(normalizedFiles, "Pipfile.lock"):
			commands = append(commands, createSetupCommand(
				"python-pipenv-install",
				"Install Python dependencies with Pipenv",
				"pipenv",
				"install",
			))
		default:
			commands = append(commands, createSetupCommand(
				"python-venv",
				"Create a Python virtual environment",
				"python",
				"-m", "venv", ".venv",
			))

			if hasFile(normalizedFiles, "requirements.txt") {
				commands = append(commands, createSetupCommand(
					"python-install-requirements",
					"Install Python dependencies from requirements.txt",
					".venv/bin/python",
					"-m", "pip", "install", "-r", "requirements.txt",
				))
			} else {
				commands = append(commands, createSetupCommand(
					"python-install-project",
					"Install the Python project in editable mode",
					".venv/bin/python",
					"-m", "pip", "install", "-e", ".",
				))
			}
		}
	}

	if hasFile(normalizedFiles, "go.mod") {
		commands = append(commands, createSetupCommand(
			"go-download",
			"Download Go module dependencies",
			"go",
			"mod", "download",
		))
	}

	if hasFile(normalizedFiles, "Cargo.toml") {
		commands = append(commands, createSetupCommand(
			"rust-fetch",
			"Fetch Rust dependencies",
			"cargo",
			"fetch",
		))
	}

	if hasFile(normalizedFiles, "pom.xml") {
		if hasFile(normalizedFiles, "mvnw") {
			commands = append(commands, createSetupCommand(
				"java-maven-wrapper-resolve",
				"Resolve Maven dependencies with the Maven wrapper",
				"./mvnw",
				"dependency:resolve",
			))
		} else {
			commands = append(commands, createSetupCommand(
				"java-maven-resolve",
				"Resolve Maven dependencies",
				"mvn",
				"dependency:resolve",
			))
		}
	}

	if hasFile(normalizedFiles, "build.gradle") || hasFile(normalizedFiles, "build.gradle.kts") {
		if hasFile(normalizedFiles, "gradlew") {
			commands = append(commands, createSetupCommand(
				"java-gradle-wrapper-build",
				"Build Gradle dependencies with the Gradle wrapper",
				"./gradlew",
				"build",
			))
		} else {
			commands = append(commands, createSetupCommand(
				"java-gradle-build",
				"Build Gradle dependencies",
				"gradle",
				"build",
			))
		}
	}

	return commands
}

func ApplySetupTemplates(commands []SetupCommand, templates map[string][]SetupCommand) []SetupCommand {
	if len(templates) == 0 {
		return commands
	}

	result := make([]SetupCommand, 0, len(commands))
	for _, command := range commands {
		if replacement, ok := templates[command.ID]; ok {
			result = append(result, replacement...)
			continue
		}

		result = append(result, command)
	}

	return result
}
