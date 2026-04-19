package cli

import (
	"bytes"
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
