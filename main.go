package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/codegame-cli-go/new/client"
	"github.com/code-game-project/codegame-cli-go/new/server"
	"github.com/spf13/pflag"
)

func main() {
	var gameName string
	pflag.StringVar(&gameName, "game-name", "", "The name of the game. (required for clients)")

	var url string
	pflag.StringVar(&url, "url", "", "The URL of the game. (required for clients)")

	var cgVersion string
	pflag.StringVar(&cgVersion, "cg-version", "", "The CodeGame protocol version of the game, e.g. 0.6 (required for clients)")

	var cgeVersion string
	pflag.StringVar(&cgeVersion, "cge-version", "", "The CGE version of the game, e.g. 0.3 (required for clients)")

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
		err = new(projectName, gameName, url, cgVersion, cgeVersion)
	default:
		err = fmt.Errorf("Unknown command: %s\n", command)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func new(projectName, gameName, url, cgVersion, cgeVersion string) error {
	projectType := strings.ToLower(pflag.Arg(1))

	var err error
	switch projectType {
	case "client":
		err = client.CreateNewClient(projectName, gameName, url, cgVersion, cgeVersion)
	case "server":
		err = server.CreateNewServer(projectName)
	default:
		err = fmt.Errorf("Unknown project type: %s\n", projectType)
	}

	return err
}
