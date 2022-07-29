package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/build"
	"github.com/code-game-project/codegame-cli-go/new/client"
	"github.com/code-game-project/codegame-cli-go/new/server"
	"github.com/code-game-project/codegame-cli-go/run"
	"github.com/code-game-project/codegame-cli/pkg/cgfile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <command> [...]\n", os.Args[0])
		os.Exit(1)
	}

	projectRoot, err := cgfile.FindProjectRoot()
	if err != nil {
		cli.Error(err.Error())
		os.Exit(1)
	}
	err = os.Chdir(projectRoot)
	if err != nil {
		cli.Error(err.Error())
		os.Exit(1)
	}

	projectName := filepath.Base(projectRoot)

	switch os.Args[1] {
	case "new":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "USAGE: %s new <client|server>\n", os.Args[0])
			os.Exit(1)
		}
		switch os.Args[2] {
		case "client":
			err = client.CreateNewClient(projectName)
		case "server":
			err = server.CreateNewServer(projectName)
		default:
			fmt.Fprintln(os.Stderr, "Unknown project type:", os.Args[2])
			os.Exit(1)
		}
	case "update":
		err = updateProject()
	case "run":
		err = runProject()
	case "build":
		err = buildProject()
	default:
		err = cli.Error("Unknown command: %s\n", os.Args[1])
	}
	if err != nil {
		os.Exit(1)
	}
}

func updateProject() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	switch config.Type {
	case "client":
		err = client.Update(config)
	case "server":
		err = server.Update()
	default:
		err = cli.Error("Unknown project type: %s\n", config.Type)
	}

	return err
}

func runProject() error {
	data, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	switch data.Type {
	case "client":
		err = run.RunClient(data.URL, args...)
	case "server":
		err = run.RunServer(args...)
	default:
		err = cli.Error("Unknown project type: %s\n", data.Type)
	}

	return err
}

func buildProject() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}
	switch config.Type {
	case "client":
		err = build.BuildClient(config.Game, config.URL)
	case "server":
		err = build.BuildServer()
	default:
		err = cli.Error("Unknown project type: %s\n", config.Type)
	}

	return err
}
