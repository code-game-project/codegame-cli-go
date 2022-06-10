package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/codegame-cli-go/new/client"
	"github.com/code-game-project/codegame-cli-go/new/server"
	"github.com/code-game-project/codegame-cli/cli"
	"github.com/spf13/pflag"
)

func main() {
	var gameName string
	pflag.StringVar(&gameName, "game-name", "", "The name of the game. (required for clients)")

	var url string
	pflag.StringVar(&url, "url", "", "The URL of the game. (required for clients)")

	var generateWrappers bool
	pflag.BoolVar(&generateWrappers, "generate-wrappers", false, "Whether to generate helper functions.")

	var libraryVersion string
	pflag.StringVar(&libraryVersion, "library-version", "latest", "The version of the Go library to use, e.g. 0.8")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nCommands:")
		fmt.Fprintln(os.Stderr, "\tnew \tCreate a new project.")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		pflag.PrintDefaults()
	}

	pflag.Parse()
	if pflag.NArg() < 2 {
		pflag.Usage()
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	projectName := filepath.Base(workingDir)

	command := strings.ToLower(pflag.Arg(0))

	switch command {
	case "new":
		err = new(projectName, gameName, url, libraryVersion, generateWrappers)
	default:
		err = cli.Error("Unknown command: %s\n", command)
	}
	if err != nil {
		os.Exit(1)
	}
}

func new(projectName, gameName, url, libraryVersion string, generateWrappers bool) error {
	projectType := strings.ToLower(pflag.Arg(1))

	var err error
	switch projectType {
	case "client":
		err = client.CreateNewClient(projectName, gameName, url, libraryVersion, generateWrappers)
	case "server":
		err = server.CreateNewServer(projectName, libraryVersion)
	default:
		err = cli.Error("Unknown project type: %s\n", projectType)
	}

	return err
}
