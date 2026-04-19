package cli

import (
	"fmt"
	"io"
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
