package server

import (
	"fmt"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/new"
	"github.com/code-game-project/codegame-cli/util/exec"
	"github.com/code-game-project/codegame-cli/util/external"
)

//go:embed templates/main.go.tmpl
var mainTemplate string

//go:embed templates/game.go.tmpl
var gameTemplate string

//go:embed templates/event_definitions.go.tmpl
var eventsTemplate string

//go:embed templates/Dockerfile.tmpl
var dockerfileTemplate string

//go:embed templates/dockerignore.tmpl
var dockerignoreTemplate string

func CreateNewServer(projectName, libraryVersion string) error {
	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	_, err = exec.Execute(true, "go", "mod", "init", module)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing go-server...")
	libraryURL, libraryTag, err := getLibraryURL(libraryVersion)
	if err != nil {
		return err
	}

	_, err = exec.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	cli.FinishLoading()

	err = createTemplate(projectName, module, libraryURL)
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

func Update(libraryVersion string) error {
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

func createTemplate(projectName, module, libraryURL string) error {
	err := executeTemplate(mainTemplate, "main.go", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	err = executeTemplate(dockerfileTemplate, "Dockerfile", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	err = executeTemplate(dockerignoreTemplate, ".dockerignore", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	packageName := strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", "")

	err = executeTemplate(gameTemplate, filepath.Join(packageName, "game.go"), projectName, libraryURL, module)
	if err != nil {
		return err
	}

	return executeTemplate(eventsTemplate, filepath.Join(packageName, "event_definitions.go"), projectName, libraryURL, module)
}

func executeTemplate(templateText, fileName, projectName, libraryURL, modulePath string) error {
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

func getLibraryURL(serverVersion string) (url string, tag string, err error) {
	if serverVersion == "latest" {
		var err error
		serverVersion, err = external.LatestGithubTag("code-game-project", "go-server")
		if err != nil {
			return "", "", err
		}
		serverVersion = strings.TrimPrefix(strings.Join(strings.Split(serverVersion, ".")[:2], "."), "v")
	}

	majorVersion := strings.Split(serverVersion, ".")[0]
	tag, err = external.GithubTagFromVersion("code-game-project", "go-server", serverVersion)
	if err != nil {
		return "", "", err
	}
	path := "github.com/code-game-project/go-server/cg"
	if majorVersion != "0" && majorVersion != "1" {
		path = fmt.Sprintf("github.com/code-game-project/go-server/v%s/cg", majorVersion)
	}

	return path, tag, nil
}
