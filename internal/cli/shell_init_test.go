package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderShellInitZshWrapsSwitch(t *testing.T) {
	t.Parallel()

	script, err := renderShellInit("zsh")
	if err != nil {
		t.Fatalf("renderShellInit: %v", err)
	}

	if !strings.Contains(script, "wtx() {") {
		t.Fatalf("script = %q", script)
	}
	if !strings.Contains(script, "command wtx path \"$@\"") {
		t.Fatalf("script = %q", script)
	}
	if !strings.Contains(script, "cd \"$target\"") {
		t.Fatalf("script = %q", script)
	}
}

func TestRenderShellInitRejectsUnsupportedShell(t *testing.T) {
	t.Parallel()

	if _, err := renderShellInit("fish"); err == nil {
		t.Fatal("expected unsupported shell error")
	}
}

func TestRunShellInitPrintsScript(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	if err := runShellInit(&stdout, "bash"); err != nil {
		t.Fatalf("runShellInit: %v", err)
	}
	if !strings.Contains(stdout.String(), "command wtx \"$@\"") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestShellWrapperFallsThroughForSwitchHelp(t *testing.T) {
	t.Parallel()

	for _, shell := range testShells(t) {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			logOutput, stdout, err := runWrappedShellCommand(t, shell, `wtx switch --help`, `#!/bin/sh
printf '%s\n' "$*" >>"$WTX_LOG"
if [ "${1-}" = "switch" ] && [ "${2-}" = "--help" ]; then
  printf 'switch help\n'
  exit 0
fi
if [ "${1-}" = "path" ]; then
  printf 'path command invoked\n'
  exit 0
fi
printf 'command:%s\n' "$*"
`)
			if err != nil {
				t.Fatalf("shell command failed: %v\nstdout=%q\nlog=%q", err, stdout, logOutput)
			}
			if strings.TrimSpace(stdout) != "switch help" {
				t.Fatalf("stdout = %q", stdout)
			}
			if strings.TrimSpace(logOutput) != "switch --help" {
				t.Fatalf("log = %q", logOutput)
			}
		})
	}
}

func TestShellWrapperFallsThroughForSwitchTargetHelp(t *testing.T) {
	t.Parallel()

	for _, shell := range testShells(t) {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			logOutput, stdout, err := runWrappedShellCommand(t, shell, `wtx switch feature/demo --help`, `#!/bin/sh
printf '%s\n' "$*" >>"$WTX_LOG"
if [ "${1-}" = "switch" ] && [ "${3-}" = "--help" ]; then
  printf 'switch target help\n'
  exit 0
fi
if [ "${1-}" = "path" ]; then
  printf 'path command invoked\n'
  exit 0
fi
printf 'command:%s\n' "$*"
`)
			if err != nil {
				t.Fatalf("shell command failed: %v\nstdout=%q\nlog=%q", err, stdout, logOutput)
			}
			if strings.TrimSpace(stdout) != "switch target help" {
				t.Fatalf("stdout = %q", stdout)
			}
			if strings.TrimSpace(logOutput) != "switch feature/demo --help" {
				t.Fatalf("log = %q", logOutput)
			}
		})
	}
}

func TestShellWrapperFallsThroughForSwitchWithoutTarget(t *testing.T) {
	t.Parallel()

	for _, shell := range testShells(t) {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			logOutput, stdout, err := runWrappedShellCommand(t, shell, `wtx switch`, `#!/bin/sh
printf '%s\n' "$*" >>"$WTX_LOG"
if [ "${1-}" = "switch" ] && [ "$#" -eq 1 ]; then
  printf 'switch usage\n'
  exit 0
fi
if [ "${1-}" = "path" ]; then
  printf 'path command invoked\n'
  exit 0
fi
`)
			if err != nil {
				t.Fatalf("shell command failed: %v\nstdout=%q\nlog=%q", err, stdout, logOutput)
			}
			if strings.TrimSpace(stdout) != "switch usage" {
				t.Fatalf("stdout = %q", stdout)
			}
			if strings.TrimSpace(logOutput) != "switch" {
				t.Fatalf("log = %q", logOutput)
			}
		})
	}
}

func TestShellWrapperBareInvocationIsSafeUnderNounset(t *testing.T) {
	t.Parallel()

	for _, shell := range testShells(t) {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			logOutput, stdout, err := runWrappedShellCommand(t, shell, `set -u; wtx`, `#!/bin/sh
printf '%s\n' "$*" >>"$WTX_LOG"
printf 'bare ok\n'
`)
			if err != nil {
				t.Fatalf("shell command failed: %v\nstdout=%q\nlog=%q", err, stdout, logOutput)
			}
			if strings.TrimSpace(stdout) != "bare ok" {
				t.Fatalf("stdout = %q", stdout)
			}
			if strings.TrimSpace(logOutput) != "" {
				t.Fatalf("log = %q", logOutput)
			}
		})
	}
}

func testShells(t *testing.T) []string {
	t.Helper()

	shells := []string{}
	for _, shell := range []string{"bash", "zsh"} {
		if _, err := exec.LookPath(shell); err == nil {
			shells = append(shells, shell)
		}
	}
	if len(shells) == 0 {
		t.Fatal("no supported shells found in PATH")
	}
	return shells
}

func runWrappedShellCommand(t *testing.T, shell, command, stubScript string) (logOutput string, stdout string, err error) {
	t.Helper()

	script, err := renderShellInit(shell)
	if err != nil {
		t.Fatalf("renderShellInit: %v", err)
	}

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "wtx.log")
	stubPath := filepath.Join(tmpDir, "wtx")
	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	shellProgram, err := exec.LookPath(shell)
	if err != nil {
		t.Fatalf("lookpath %s: %v", shell, err)
	}

	cmd := exec.Command(shellProgram, "-c", script+"\n"+command)
	cmd.Env = append(os.Environ(),
		"PATH="+tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"WTX_LOG="+logPath,
	)
	output, runErr := cmd.CombinedOutput()

	logBytes, readErr := os.ReadFile(logPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("read log: %v", readErr)
	}

	return string(logBytes), string(output), runErr
}
