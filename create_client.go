package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"
)

//go:embed templates/new/client/main.go.tmpl
var clientMainTemplate string

//go:embed templates/new/client/game.go.tmpl
var clientGameTemplate string

//go:embed templates/new/client/events.go.tmpl
var clientEventsTemplate string

//go:embed templates/new/gitignore.tmpl
var gitignoreTemplate string

func CreateClient() error {
	fmt.Println("create client")
	return nil
}

func createClientTemplate(modulePath, gameName, libraryURL string, eventNames, commandNames []string) error {
	return execClientTemplate(modulePath, gameName, libraryURL, eventNames, commandNames, false)
}

func execClientTemplate(modulePath, gameName, libraryURL string, eventNames, commandNames []string, update bool) error {
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")
	gameDir := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	if update {
		// cli.Warn("This action will ERASE and regenerate ALL files in '%s/'.\nYou will have to manually update your code to work with the new version.", gameDir)
		// ok, err := cli.YesNo("Continue?", false)
		// if err != nil || !ok {
		// return cli.ErrCanceled
		// }
		os.RemoveAll(gameDir)
	} else {
		// cli.Warn("DO NOT EDIT the `%s/` directory inside of the project. ALL CHANGES WILL BE LOST when running `codegame update`.", gameDir)
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
