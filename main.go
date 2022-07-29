package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/Bananenpro/pflag"
	"github.com/code-game-project/codegame-cli-go/build"
	"github.com/code-game-project/codegame-cli-go/new/client"
	"github.com/code-game-project/codegame-cli-go/new/server"
	"github.com/code-game-project/codegame-cli-go/run"
	"github.com/code-game-project/codegame-cli/util/cgfile"
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nCommands:")
		fmt.Fprintln(os.Stderr, "\tnew \tCreate a new project.")
		fmt.Fprintln(os.Stderr, "\tupdate \tUpdate the current project.")
		fmt.Fprintln(os.Stderr, "\trun \tRun the current project.")
		fmt.Fprintln(os.Stderr, "\tbuild \tBuild the current project.")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		pflag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "This program expects to be executed inside of the project directory.")
	}

	if len(os.Args) < 2 {
		pflag.Usage()
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	projectName := filepath.Base(workingDir)

	command := strings.ToLower(os.Args[1])

	switch command {
	case "new":
		err = newProject(projectName)
	case "update":
		err = updateProject()
	case "run":
		err = runProject()
	case "build":
		err = buildProject()
	default:
		err = cli.Error("Unknown command: %s\n", command)
	}
	if err != nil {
		os.Exit(1)
	}
}

func newProject(projectName string) error {
	flagSet := pflag.NewFlagSet("new", pflag.ExitOnError)

	var gameName string
	flagSet.StringVar(&gameName, "game-name", "", "The name of the game. (required for `new client`)")

	var url string
	flagSet.StringVar(&url, "url", "", "The URL of the game. (required for `new client`)")

	var libraryVersion string
	flagSet.StringVar(&libraryVersion, "library-version", "latest", "The version of the Go library to use, e.g. 0.8")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s new <client|server>\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[2:])

	if flagSet.NArg() == 0 {
		flagSet.Usage()
		os.Exit(1)
	}
	projectType := flagSet.Arg(0)

	var err error
	switch projectType {
	case "client":
		err = client.CreateNewClient(projectName, gameName, url, libraryVersion)
	case "server":
		err = server.CreateNewServer(projectName, libraryVersion)
	default:
		err = cli.Error("Unknown project type: %s\n", projectType)
	}

	return err
}

func updateProject() error {
	flagSet := pflag.NewFlagSet("new", pflag.ExitOnError)

	var libraryVersion string
	flagSet.StringVar(&libraryVersion, "library-version", "latest", "The version of the Go library to use, e.g. 0.8")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s update\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[2:])

	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	switch config.Type {
	case "client":
		err = client.Update(libraryVersion, config)
	case "server":
		err = server.Update(libraryVersion)
	default:
		err = cli.Error("Unknown project type: %s\n", config.Type)
	}

	return err
}

func runProject() error {
	flagSet := pflag.NewFlagSet("run", pflag.ExitOnError)
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags:           true,
		PassUnknownFlagsToArgs: true,
	}

	var overrideURL string
	flagSet.StringVar(&overrideURL, "override-url", "", "The URL of the game. (required for `new client`)")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s run [...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[2:])

	projectRoot, err := cgfile.FindProjectRootRelative()
	if err != nil {
		return err
	}

	data, err := cgfile.LoadCodeGameFile(projectRoot)
	if err != nil {
		return err
	}

	url := data.URL
	if overrideURL != "" {
		url = overrideURL
	}

	switch data.Type {
	case "client":
		err = run.RunClient(projectRoot, url, flagSet.Args()...)
	case "server":
		err = run.RunServer(projectRoot, flagSet.Args()...)
	default:
		err = cli.Error("Unknown project type: %s\n", data.Type)
	}

	return err
}

func buildProject() error {
	flagSet := pflag.NewFlagSet("build", pflag.ExitOnError)

	var output string
	flagSet.StringVarP(&output, "output", "o", "", "The name of the output file.")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s build [...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nOptions:")
		fmt.Fprintf(os.Stderr, "`update` must be executed at the root of the project.")
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[2:])

	projectRoot, err := cgfile.FindProjectRootRelative()
	if err != nil {
		return err
	}

	data, err := cgfile.LoadCodeGameFile(projectRoot)
	if err != nil {
		return err
	}

	switch data.Type {
	case "client":
		err = build.BuildClient(projectRoot, data.Game, output, data.URL)
	case "server":
		err = build.BuildServer(projectRoot, output)
	default:
		err = cli.Error("Unknown project type: %s\n", data.Type)
	}

	return err
}
