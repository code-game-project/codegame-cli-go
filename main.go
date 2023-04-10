package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bananenpro/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <action> [...]\n", os.Args[0])
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	projectName := filepath.Base(workingDir)

	switch os.Args[1] {
	case "info":
		err = Info()
	case "create":
		err = Create(projectName)
	default:
		fmt.Fprintln(os.Stderr, "unsupported action:", os.Args[1])
		os.Exit(1)
	}
	if err != nil {
		if err != cli.ErrCanceled {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(1)
	}
}
