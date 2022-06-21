package client

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/new"
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

	err = createClientTemplate(projectName, module, gameName, serverURL, libraryURL, generateWrappers, eventNames)
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

func createClientTemplate(projectName, modulePath, gameName, serverURL, libraryURL string, wrappers bool, eventNames []string) error {
	if !wrappers {
		return execClientMainTemplate(projectName, serverURL, libraryURL)
	}

	return execClientWrappersTemplate(projectName, modulePath, gameName, serverURL, libraryURL, eventNames)
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

func execClientMainTemplate(projectName, serverURL, libraryURL string) error {
	type data struct {
		URL        string
		LibraryURL string
	}

	return new.ExecTemplate(mainTemplate, "main.go", data{
		URL:        serverURL,
		LibraryURL: libraryURL,
	})
}

func execClientWrappersTemplate(projectName, modulePath, gameName, serverURL, libraryURL string, eventNames []string) error {
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	gameDir := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

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

	err := new.ExecTemplate(wrapperMainTemplate, filepath.Join("main.go"), data)
	if err != nil {
		return err
	}

	err = new.ExecTemplate(wrapperGameTemplate, filepath.Join(gameDir, "game.go"), data)
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
