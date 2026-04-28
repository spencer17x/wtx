package core

import (
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

func normalizeSetupDirectory(directory string) string {
	normalized := strings.ReplaceAll(directory, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimRight(normalized, "/")
	if normalized == "." {
		return ""
	}
	return normalized
}

func fileNameInSetupDirectory(file string, directory string) (string, bool) {
	normalizedFile := strings.ReplaceAll(file, "\\", "/")
	normalizedFile = strings.TrimPrefix(normalizedFile, "./")
	normalizedDirectory := normalizeSetupDirectory(directory)

	if normalizedDirectory == "" {
		if normalizedFile == "" || strings.Contains(normalizedFile, "/") {
			return "", false
		}
		return normalizedFile, true
	}

	prefix := normalizedDirectory + "/"
	if !strings.HasPrefix(normalizedFile, prefix) {
		return "", false
	}

	relativeFile := strings.TrimPrefix(normalizedFile, prefix)
	if relativeFile == "" || strings.Contains(relativeFile, "/") {
		return "", false
	}

	return relativeFile, true
}

func normalizeFileSetForDirectory(files []string, directory string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(files))
	for _, file := range files {
		fileName, ok := fileNameInSetupDirectory(file, directory)
		if !ok {
			continue
		}
		normalized[fileName] = struct{}{}
	}
	return normalized
}

func hasFile(files map[string]struct{}, expected string) bool {
	_, ok := files[expected]
	return ok
}

func detectSetupCommands(normalizedFiles map[string]struct{}) []SetupCommand {
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

func DetectSetupCommands(files []string) []SetupCommand {
	return DetectSetupCommandsForDirectory(files, "")
}

func DetectSetupCommandsForDirectory(files []string, directory string) []SetupCommand {
	workingDirectory := normalizeSetupDirectory(directory)
	commands := detectSetupCommands(normalizeFileSetForDirectory(files, workingDirectory))
	for index := range commands {
		commands[index].WorkingDirectory = workingDirectory
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
			for _, replacementCommand := range replacement {
				if replacementCommand.WorkingDirectory == "" {
					replacementCommand.WorkingDirectory = command.WorkingDirectory
				}
				result = append(result, replacementCommand)
			}
			continue
		}

		result = append(result, command)
	}

	return result
}
