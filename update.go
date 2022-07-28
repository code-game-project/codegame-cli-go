package main

import (
	"fmt"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli/pkg/cgfile"
	"github.com/code-game-project/codegame-cli/pkg/cggenevents"
	"github.com/code-game-project/codegame-cli/pkg/exec"
	"github.com/code-game-project/codegame-cli/pkg/external"
	"github.com/code-game-project/codegame-cli/pkg/modules"
	"github.com/code-game-project/codegame-cli/pkg/server"
)

func Update() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.UpdateData]()
	if err != nil {
		return err
	}
	switch config.Type {
	case "client":
		return updateClient(data.LibraryVersion, config)
	case "server":
		return updateServer(data.LibraryVersion)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func updateClient(libraryVersion string, config *cgfile.CodeGameFileData) error {
	baseURL := external.BaseURL("http", external.IsTLS(config.URL), config.URL)
	api, err := server.NewAPI(baseURL)
	libraryURL, libraryTag, err := getClientLibraryURL(libraryVersion)
	if err != nil {
		return err
	}

	cge, err := api.GetCGEFile()
	if err != nil {
		return err
	}
	cgeVersion, err := cggenevents.ParseCGEVersion(cge)
	if err != nil {
		return err
	}

	eventNames, commandNames, err := cggenevents.GetEventNames(config.URL, cgeVersion)
	if err != nil {
		return err
	}

	module, err := GetGoModuleName("")
	if err != nil {
		return err
	}

	err = updateClientTemplate(module, config.Game, config.URL, libraryURL, eventNames, commandNames)
	if err != nil {
		return err
	}

	cli.BeginLoading("Updating dependencies...")
	_, err = exec.Execute(true, "go", "get", "-u", "./...")
	if err != nil {
		return err
	}
	_, err = exec.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	_, err = exec.Execute(true, "go", "mod", "tidy")
	if err != nil {
		return err
	}
	cli.FinishLoading()
	return nil
}

func updateClientTemplate(modulePath, gameName, serverURL, libraryURL string, eventNames, commandNames []string) error {
	return execClientTemplate(modulePath, gameName, serverURL, libraryURL, eventNames, commandNames, true)
}

func updateServer(libraryVersion string) error {
	cli.Warn("This update might include breaking changes. You will have to manually update your code to work with the new version.")
	ok, err := cli.YesNo("Continue?", false)
	if err != nil || !ok {
		return cli.ErrCanceled
	}

	cli.BeginLoading("Updating dependencies...")

	_, err = exec.Execute(true, "go", "get", "-u", "./...")
	if err != nil {
		return err
	}

	cli.FinishLoading()
	return nil
}
