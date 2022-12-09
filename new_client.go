package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/server"
)

//go:embed templates/new/client/main.go.tmpl
var clientMainTemplate string

//go:embed templates/new/client/game.go.tmpl
var clientGameTemplate string

//go:embed templates/new/client/events.go.tmpl
var clientEventsTemplate string

//go:embed templates/new/gitignore.tmpl
var gitignoreTemplate string

func CreateNewClient() error {
	data, err := modules.ReadCommandConfig[modules.NewClientData]()
	if err != nil {
		return err
	}

	api, err := server.NewAPI(data.URL)
	if err != nil {
		return err
	}

	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	_, err = exec.Execute(true, "go", "mod", "init", module)
	if err != nil {
		return err
	}

	libraryURL, libraryTag, err := getClientLibraryURL(data.LibraryVersion)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing go-client...")
	_, err = exec.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	cli.FinishLoading()

	cge, err := api.GetCGEFile()
	if err != nil {
		return err
	}
	cgeVersion, err := cggenevents.ParseCGEVersion(cge)
	if err != nil {
		return err
	}

	eventNames, commandNames, err := cggenevents.GetEventNames(api.BaseURL(), cgeVersion)
	if err != nil {
		return err
	}

	err = createClientTemplate(module, data.Name, libraryURL, eventNames, commandNames)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing dependencies...")

	_, err = exec.Execute(true, "go", "mod", "tidy")
	if err != nil {
		return err
	}

	cli.FinishLoading()
	return nil
}

func createClientTemplate(modulePath, gameName, libraryURL string, eventNames, commandNames []string) error {
	return execClientTemplate(modulePath, gameName, libraryURL, eventNames, commandNames, false)
}

func getClientLibraryURL(clientVersion string) (url string, tag string, err error) {
	if clientVersion == "latest" {
		clientVersion, err = external.LatestGithubTag("code-game-project", "go-client")
		if err != nil {
			return "", "", err
		}
		clientVersion = strings.TrimPrefix(strings.Join(strings.Split(clientVersion, ".")[:2], "."), "v")
	}
	majorVersion := strings.Split(clientVersion, ".")[0]
	tag, err = external.GithubTagFromVersion("code-game-project", "go-client", clientVersion)
	if err != nil {
		return "", "", err
	}
	path := "github.com/code-game-project/go-client/cg"
	if majorVersion != "0" && majorVersion != "1" {
		path = fmt.Sprintf("github.com/code-game-project/go-client/v%s/cg", majorVersion)
	}

	return path, tag, nil
}

func execClientTemplate(modulePath, gameName, libraryURL string, eventNames, commandNames []string, update bool) error {
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")
	gameDir := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	if update {
		cli.Warn("This action will ERASE and regenerate ALL files in '%s/'.\nYou will have to manually update your code to work with the new version.", gameDir)
		ok, err := cli.YesNo("Continue?", false)
		if err != nil || !ok {
			return cli.ErrCanceled
		}
		os.RemoveAll(gameDir)
	} else {
		cli.Warn("DO NOT EDIT the `%s/` directory inside of the project. ALL CHANGES WILL BE LOST when running `codegame update`.", gameDir)
	}

	type event struct {
		Name       string
		PascalName string
	}

	events := make([]event, len(eventNames))
	for i, e := range eventNames {
		pascal := strings.ReplaceAll(e, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		events[i] = event{
			Name:       e,
			PascalName: pascal,
		}
	}

	commands := make([]event, len(commandNames))
	for i, c := range commandNames {
		pascal := strings.ReplaceAll(c, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		commands[i] = event{
			Name:       c,
			PascalName: pascal,
		}
	}

	data := struct {
		LibraryURL  string
		PackageName string
		ModulePath  string
		Events      []event
		Commands    []event
	}{
		LibraryURL:  libraryURL,
		PackageName: gamePackageName,
		ModulePath:  modulePath,
		Events:      events,
		Commands:    commands,
	}

	if !update {
		err := ExecTemplate(clientMainTemplate, "main.go", data)
		if err != nil {
			return err
		}

		err = ExecTemplate(gitignoreTemplate, ".gitignore", data)
		if err != nil {
			return err
		}
	}

	err := ExecTemplate(clientGameTemplate, filepath.Join(gameDir, "game.go"), data)
	if err != nil {
		return err
	}

	err = ExecTemplate(clientEventsTemplate, filepath.Join(gameDir, "events.go"), data)
	if err != nil {
		return err
	}

	return nil
}
