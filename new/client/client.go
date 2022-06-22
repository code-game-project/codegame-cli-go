package client

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/new"
	"github.com/code-game-project/codegame-cli-go/util"
	"github.com/code-game-project/codegame-cli/util/cgfile"
	"github.com/code-game-project/codegame-cli/util/cggenevents"
	"github.com/code-game-project/codegame-cli/util/exec"
	"github.com/code-game-project/codegame-cli/util/external"
)

//go:embed templates/main.go.tmpl
var mainTemplate string

//go:embed templates/wrappers/main.go.tmpl
var wrapperMainTemplate string

//go:embed templates/wrappers/game.go.tmpl
var wrapperGameTemplate string

//go:embed templates/wrappers/events.go.tmpl
var wrapperEventsTemplate string

func CreateNewClient(projectName, gameName, serverURL, libraryVersion string, generateWrappers bool) error {
	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	_, err = exec.Execute(true, "go", "mod", "init", module)
	if err != nil {
		return err
	}

	libraryURL, libraryTag, err := getClientLibraryURL(libraryVersion)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing go-client...")
	_, err = exec.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	cli.FinishLoading()

	var eventNames []string
	if generateWrappers {
		cgeVersion, err := cggenevents.GetCGEVersion(baseURL(serverURL, isSSL(serverURL)))
		if err != nil {
			return err
		}

		eventNames, err = cggenevents.GetEventNames(baseURL(serverURL, isSSL(serverURL)), cgeVersion)
		if err != nil {
			return err
		}
	}

	err = createClientTemplate(module, gameName, serverURL, libraryURL, generateWrappers, eventNames)
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

func Update(libraryVersion string, config *cgfile.CodeGameFileData) error {
	libraryURL, libraryTag, err := getClientLibraryURL(libraryVersion)
	if err != nil {
		return err
	}

	url := baseURL(config.URL, isSSL(config.URL))

	var eventNames []string
	cgeVersion, err := cggenevents.GetCGEVersion(url)
	if err != nil {
		return err
	}

	eventNames, err = cggenevents.GetEventNames(url, cgeVersion)
	if err != nil {
		return err
	}

	module, err := util.GetModuleName("")
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

func createClientTemplate(modulePath, gameName, serverURL, libraryURL string, wrappers bool, eventNames []string) error {
	if !wrappers {
		return execClientMainTemplate(serverURL, libraryURL)
	}

	return execClientWrappersTemplate(modulePath, gameName, serverURL, libraryURL, eventNames, false)
}

func updateClientTemplate(modulePath, gameName, serverURL, libraryURL string, eventNames []string) error {
	return execClientWrappersTemplate(modulePath, gameName, serverURL, libraryURL, eventNames, true)
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

func execClientMainTemplate(serverURL, libraryURL string) error {
	type data struct {
		URL        string
		LibraryURL string
	}

	return new.ExecTemplate(mainTemplate, "main.go", data{
		URL:        serverURL,
		LibraryURL: libraryURL,
	})
}

func execClientWrappersTemplate(modulePath, gameName, serverURL, libraryURL string, eventNames []string, update bool) error {
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
		err := new.ExecTemplate(wrapperMainTemplate, filepath.Join("main.go"), data)
		if err != nil {
			return err
		}
	}

	err := new.ExecTemplate(wrapperGameTemplate, filepath.Join(gameDir, "game.go"), data)
	if err != nil {
		return err
	}

	err = new.ExecTemplate(wrapperEventsTemplate, filepath.Join(gameDir, "events.go"), data)
	if err != nil {
		return err
	}

	return nil
}

func baseURL(domain string, ssl bool) string {
	if ssl {
		return "https://" + domain
	} else {
		return "http://" + domain
	}
}

func isSSL(domain string) bool {
	res, err := http.Get("https://" + domain)
	if err == nil {
		res.Body.Close()
		return true
	}
	return false
}
