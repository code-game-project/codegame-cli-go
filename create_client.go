package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "embed"

	"github.com/code-game-project/cge-parser/adapter"
	"github.com/code-game-project/cli-module/module"
	"github.com/code-game-project/cli-utils/cli"
	"github.com/code-game-project/cli-utils/exec"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/templates"
)

//go:embed templates/new/client/main.go.tmpl
var clientMainTemplate string

//go:embed templates/new/client/game.go.tmpl
var clientGameTemplate string

//go:embed templates/new/client/events.go.tmpl
var clientEventsTemplate string

//go:embed templates/new/gitignore.tmpl
var gitignoreTemplate string

var simpleModPathRegex = regexp.MustCompile("[a-z]+[a-z0-9]*")

func CreateClient(data *modules.ActionCreateData, modulePath string) error {
	libraryURL, libraryTag, err := getLibraryURL("go-client", data.LibraryVersion)
	if err != nil {
		return fmt.Errorf("determine client library details: %w", err)
	}

	feedback.Info(FeedbackPkg, "Installing go-client %s...", libraryTag)
	err = exec.ExecuteDimmed("go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return fmt.Errorf("install go-client: %w", err)
	}

	feedback.Info(FeedbackPkg, "Loading CGE data...")
	cgeData, err := module.LoadCGEData(data.GetUrl())
	if err != nil {
		return fmt.Errorf("load cge data: %w", err)
	}

	feedback.Info(FeedbackPkg, "Generating template...")
	err = createClientTemplate(modulePath, data.GameName, libraryURL, cgeData)
	if err != nil {
		return fmt.Errorf("generate template: %w", err)
	}

	feedback.Info(FeedbackPkg, "Installing dependencies...")
	err = exec.ExecuteDimmed("go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("install dependencies: %w", err)
	}

	return nil
}

func createClientTemplate(modulePath, gameName, libraryURL string, cgeData adapter.ParserResponse) error {
	return execClientTemplate(modulePath, gameName, libraryURL, cgeData, false)
}

func execClientTemplate(modulePath, gameName, libraryURL string, cgeData adapter.ParserResponse, update bool) error {
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")
	gameDir := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	if update {
		feedback.Warn(FeedbackPkg, "This action will ERASE and regenerate ALL files in '%s/'.\nYou will have to manually update your code to work with the new version.", gameDir)
		ok := cli.YesNo("Continue?", false)
		if !ok {
			os.Exit(2)
		}
		os.RemoveAll(gameDir)
	} else {
		feedback.Warn(FeedbackPkg, "DO NOT EDIT the `%s/` directory inside of the project. ALL CHANGES WILL BE LOST when running `codegame update`.", gameDir)
	}

	type event struct {
		Name       string
		PascalName string
	}

	eventNames := make([]event, len(cgeData.Events))
	for i, e := range cgeData.Events {
		pascal := strings.ReplaceAll(e.Name, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		eventNames[i] = event{
			Name:       e.Name,
			PascalName: pascal,
		}
	}

	commandNames := make([]event, len(cgeData.Commands))
	for i, c := range cgeData.Commands {
		pascal := strings.ReplaceAll(c.Name, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		commandNames[i] = event{
			Name:       c.Name,
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
		Events:      eventNames,
		Commands:    commandNames,
	}

	if !update {
		err := templates.Execute(clientMainTemplate, "main.go", data)
		if err != nil {
			return err
		}

		err = templates.Execute(gitignoreTemplate, ".gitignore", data)
		if err != nil {
			return err
		}
	}

	err := templates.Execute(clientGameTemplate, filepath.Join(gameDir, "game.go"), data)
	if err != nil {
		return err
	}

	err = templates.Execute(clientEventsTemplate, filepath.Join(gameDir, "events.go"), data)
	if err != nil {
		return err
	}

	eventDefinitionsFile, err := os.Create(filepath.Join(gameDir, "definitions.go"))
	if err != nil {
		return fmt.Errorf("create definitions.go: %w", err)
	}
	generateEventDefinitions(eventDefinitionsFile, gamePackageName, libraryURL, cgeData)
	eventDefinitionsFile.Close()
	return nil
}
