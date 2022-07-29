package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/new"
	"github.com/code-game-project/codegame-cli-go/util"
	"github.com/code-game-project/codegame-cli/pkg/cgfile"
	"github.com/code-game-project/codegame-cli/pkg/cggenevents"
	"github.com/code-game-project/codegame-cli/pkg/exec"
	"github.com/code-game-project/codegame-cli/pkg/external"
	"github.com/code-game-project/codegame-cli/pkg/modules"
	"github.com/code-game-project/codegame-cli/pkg/server"
)

//go:embed templates/main.go.tmpl
var mainTemplate string

//go:embed templates/game.go.tmpl
var gameTemplate string

//go:embed templates/events.go.tmpl
var eventsTemplate string

func CreateNewClient(projectName string) error {
	data, err := modules.ReadCommandConfig[modules.NewClientData]()
	if err != nil {
		return err
	}
	data.URL = external.TrimURL(data.URL)

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

	eventNames, _, err := cggenevents.GetEventNames(external.BaseURL("http", external.IsTLS(data.URL), data.URL), cgeVersion)
	if err != nil {
		return err
	}

	err = createClientTemplate(module, data.Name, data.URL, libraryURL, eventNames)
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

func Update(config *cgfile.CodeGameFileData) error {
	data, err := modules.ReadCommandConfig[modules.NewServerData]()
	if err != nil {
		return err
	}

	api, err := server.NewAPI(config.URL)
	if err != nil {
		return err
	}

	libraryURL, libraryTag, err := getClientLibraryURL(data.LibraryVersion)
	if err != nil {
		return err
	}

	url := external.BaseURL("http", external.IsTLS(config.URL), config.URL)

	cge, err := api.GetCGEFile()
	if err != nil {
		return err
	}

	cgeVersion, err := cggenevents.ParseCGEVersion(cge)
	if err != nil {
		return err
	}

	eventNames, _, err := cggenevents.GetEventNames(url, cgeVersion)
	if err != nil {
		return err
	}

	module, err := util.GetModuleName()
	if err != nil {
		return err
	}

	err = updateClientTemplate(module, config.Game, config.URL, libraryURL, eventNames)
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

func createClientTemplate(modulePath, gameName, serverURL, libraryURL string, eventNames []string) error {
	return execClientTemplate(modulePath, gameName, serverURL, libraryURL, eventNames, false)
}

func updateClientTemplate(modulePath, gameName, serverURL, libraryURL string, eventNames []string) error {
	return execClientTemplate(modulePath, gameName, serverURL, libraryURL, eventNames, true)
}

func getClientLibraryURL(clientVersion string) (url string, tag string, err error) {
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

func execClientTemplate(modulePath, gameName, serverURL, libraryURL string, eventNames []string, update bool) error {
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
		cli.Warn("DO NOT EDIT the `%s/` directory. ALL CHANGES WILL BE LOST when running `codegame update`.", gameDir)
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

	data := struct {
		URL         string
		LibraryURL  string
		PackageName string
		ModulePath  string
		Events      []event
	}{
		URL:         serverURL,
		LibraryURL:  libraryURL,
		PackageName: gamePackageName,
		ModulePath:  modulePath,
		Events:      events,
	}

	if !update {
		err := new.ExecTemplate(mainTemplate, filepath.Join("main.go"), data)
		if err != nil {
			return err
		}
	}

	err := new.ExecTemplate(gameTemplate, filepath.Join(gameDir, "game.go"), data)
	if err != nil {
		return err
	}

	err = new.ExecTemplate(eventsTemplate, filepath.Join(gameDir, "events.go"), data)
	if err != nil {
		return err
	}

	return nil
}
