package client

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/code-game-project/codegame-cli-go/new"
	"github.com/code-game-project/codegame-cli/cli"
	"github.com/code-game-project/codegame-cli/util"
)

//go:embed templates/main.go.tmpl
var goClientMainTemplate string

//go:embed templates/wrappers/main.go.tmpl
var goClientWrapperMainTemplate string

//go:embed templates/wrappers/game.go.tmpl
var goClientWrapperGameTemplate string

//go:embed templates/wrappers/events.go.tmpl
var goClientWrapperEventsTemplate string

func CreateNewClient(projectName, gameName, serverURL, libraryVersion string, generateWrappers bool) error {
	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	_, err = util.Execute(true, "go", "mod", "init", module)
	if err != nil {
		return err
	}

	libraryURL, libraryTag, err := getGoClientLibraryURL(libraryVersion)
	if err != nil {
		return err
	}

	cli.Begin("Installing correct go-client version...")
	_, err = util.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	cli.Finish()

	cgeVersion, err := util.GetCGEVersion(baseURL(serverURL, isSSL(serverURL)))
	if err != nil {
		return err
	}

	cli.Begin("Creating project template...")
	err = createGoClientTemplate(projectName, module, gameName, serverURL, libraryURL, cgeVersion, generateWrappers)
	if err != nil {
		return err
	}
	cli.Finish()

	cli.Begin("Installing dependencies...")

	_, err = util.Execute(true, "go", "mod", "tidy")
	if err != nil {
		return err
	}

	cli.Finish()

	cli.Begin("Organizing imports...")

	if !util.IsInstalled("goimports") {
		cli.Warn("Failed to organize import statements: 'goimports' is not installed!")
		return nil
	}
	util.Execute(true, "goimports", "-w", "main.go")

	cli.Finish()

	return nil
}

func createGoClientTemplate(projectName, modulePath, gameName, serverURL, libraryURL, cgeVersion string, wrappers bool) error {
	if !wrappers {
		return execGoClientMainTemplate(projectName, serverURL, libraryURL)
	}

	return execGoClientWrappersTemplate(projectName, modulePath, gameName, serverURL, libraryURL, cgeVersion)
}

func getGoClientLibraryURL(clientVersion string) (url string, tag string, err error) {
	if clientVersion == "latest" {
		var err error
		clientVersion, err = util.LatestGithubTag("code-game-project", "go-client")
		if err != nil {
			return "", "", err
		}
		clientVersion = strings.TrimPrefix(strings.Join(strings.Split(clientVersion, ".")[:2], "."), "v")
	}

	majorVersion := strings.Split(clientVersion, ".")[0]
	tag, err = util.GithubTagFromVersion("code-game-project", "go-client", clientVersion)
	if err != nil {
		return "", "", err
	}
	path := "github.com/code-game-project/go-client/cg"
	if majorVersion != "0" && majorVersion != "1" {
		path = fmt.Sprintf("github.com/code-game-project/go-client/v%s/cg", majorVersion)
	}

	return path, tag, nil
}

func execGoClientMainTemplate(projectName, serverURL, libraryURL string) error {
	type data struct {
		URL        string
		LibraryURL string
	}

	return new.ExecTemplate(goClientMainTemplate, "main.go", data{
		URL:        serverURL,
		LibraryURL: libraryURL,
	})
}

func execGoClientWrappersTemplate(projectName, modulePath, gameName, serverURL, libraryURL, cgeVersion string) error {
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	gameDir := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	eventNames, err := util.GetEventNames(baseURL(serverURL, isSSL(serverURL)), cgeVersion)
	if err != nil {
		return err
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

	err = new.ExecTemplate(goClientWrapperMainTemplate, filepath.Join("main.go"), data)
	if err != nil {
		return err
	}

	err = new.ExecTemplate(goClientWrapperGameTemplate, filepath.Join(gameDir, "game.go"), data)
	if err != nil {
		return err
	}

	err = new.ExecTemplate(goClientWrapperEventsTemplate, filepath.Join(gameDir, "events.go"), data)
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
