package cli

import (
	"fmt"
	"io"
)

func renderShellInit(shell string) (string, error) {
	switch shell {
	case "zsh", "bash":
		return `wtx() {
  if [ "$1" = "switch" ]; then
    shift
    local target
    target="$(command wtx path "$@")" || return $?
    cd "$target"
  else
    command wtx "$@"
  fi
}
`, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

func runShellInit(stdout io.Writer, shell string) error {
	script, err := renderShellInit(shell)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(stdout, script)
	return err
}
