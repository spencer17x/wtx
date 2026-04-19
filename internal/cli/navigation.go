package cli

import (
	"fmt"
	"io"
	"strings"
)

type resolvePathFunc func(branchName string) (string, error)

func runPath(stdout io.Writer, branchName string, resolve resolvePathFunc) error {
	path, err := resolve(branchName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, path)
	return err
}

func runSwitch(stdout io.Writer, branchName string, resolve resolvePathFunc) error {
	path, err := resolve(branchName)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(stdout, path); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "shell integration not enabled; use: cd \"$(wtx path %s)\"\n", shellQuote(branchName))
	return err
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
