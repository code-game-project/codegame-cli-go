package server

import (
	"fmt"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/code-game-project/codegame-cli-go/new"
	"github.com/code-game-project/codegame-cli/cli"
	"github.com/code-game-project/codegame-cli/util"
)

//go:embed templates/main.go.tmpl
var goServerMainTemplate string

//go:embed templates/game.go.tmpl
var goServerGameTemplate string

//go:embed templates/event_definitions.go.tmpl
var goServerEventsTemplate string

func CreateNewServer(projectName, libraryVersion string) error {
	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	_, err = util.Execute(true, "go", "mod", "init", module)
	if err != nil {
		return err
	}

	cli.Begin("Installing correct go-server version...")
	libraryURL, libraryTag, err := getServerLibraryURL(libraryVersion)
	if err != nil {
		return err
	}

	_, err = util.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	cli.Finish()

	cli.Begin("Creating project template...")
	err = createGoServerTemplate(projectName, module, libraryURL)
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

	packageDir := strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", "")
	util.Execute(true, "goimports", "-w", filepath.Join(packageDir, "game.go"))

	cli.Finish()

	return nil
}

func createGoServerTemplate(projectName, module, libraryURL string) error {
	err := executeGoServerTemplate(goServerMainTemplate, "main.go", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	packageName := strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", "")

	err = executeGoServerTemplate(goServerGameTemplate, filepath.Join(packageName, "game.go"), projectName, libraryURL, module)
	if err != nil {
		return err
	}

	return executeGoServerTemplate(goServerEventsTemplate, filepath.Join(packageName, "event_definitions.go"), projectName, libraryURL, module)
}

func executeGoServerTemplate(templateText, fileName, projectName, libraryURL, modulePath string) error {
	type data struct {
		Name        string
		PackageName string
		LibraryURL  string
		ModulePath  string
	}

	return new.ExecTemplate(templateText, fileName, data{
		Name:        projectName,
		PackageName: strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", ""),
		LibraryURL:  libraryURL,
		ModulePath:  modulePath,
	})
}

func getServerLibraryURL(serverVersion string) (url string, tag string, err error) {
	if serverVersion == "latest" {
		var err error
		serverVersion, err = util.LatestGithubTag("code-game-project", "go-server")
		if err != nil {
			return "", "", err
		}
		serverVersion = strings.TrimPrefix(strings.Join(strings.Split(serverVersion, ".")[:2], "."), "v")
	}

	majorVersion := strings.Split(serverVersion, ".")[0]
	tag, err = util.GithubTagFromVersion("code-game-project", "go-server", serverVersion)
	if err != nil {
		return "", "", err
	}
	path := "github.com/code-game-project/go-server/cg"
	if majorVersion != "0" && majorVersion != "1" {
		path = fmt.Sprintf("github.com/code-game-project/go-server/v%s/cg", majorVersion)
	}

	return path, tag, nil
}
