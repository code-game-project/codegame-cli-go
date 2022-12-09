package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"

	_ "embed"
)

//go:embed templates/new/server/main.go.tmpl
var serverMainTemplate string

//go:embed templates/new/server/game.go.tmpl
var serverGameTemplate string

//go:embed templates/new/server/event_definitions.go.tmpl
var serverEventsTemplate string

//go:embed templates/new/server/Dockerfile.tmpl
var serverDockerfileTemplate string

//go:embed templates/new/server/dockerignore.tmpl
var serverDockerignoreTemplate string

func CreateNewServer(projectName string) error {
	data, err := modules.ReadCommandConfig[modules.NewServerData]()

	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	_, err = exec.Execute(true, "go", "mod", "init", module)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing go-server...")
	libraryURL, libraryTag, err := getServerLibraryURL(data.LibraryVersion)
	if err != nil {
		return err
	}

	_, err = exec.Execute(true, "go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return err
	}
	cli.FinishLoading()

	err = createServerTemplate(projectName, module, libraryURL)
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

func createServerTemplate(projectName, module, libraryURL string) error {
	err := executeServerTemplate(serverMainTemplate, "main.go", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	err = executeServerTemplate(gitignoreTemplate, ".gitignore", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	err = executeServerTemplate(serverDockerfileTemplate, "Dockerfile", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	err = executeServerTemplate(serverDockerignoreTemplate, ".dockerignore", projectName, libraryURL, module)
	if err != nil {
		return err
	}

	packageName := strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", "")

	err = executeServerTemplate(serverGameTemplate, filepath.Join(packageName, "game.go"), projectName, libraryURL, module)
	if err != nil {
		return err
	}

	return executeServerTemplate(serverEventsTemplate, filepath.Join(packageName, "event_definitions.go"), projectName, libraryURL, module)
}

func executeServerTemplate(templateText, fileName, projectName, libraryURL, modulePath string) error {
	type data struct {
		Name        string
		PackageName string
		LibraryURL  string
		ModulePath  string
	}

	return ExecTemplate(templateText, fileName, data{
		Name:        projectName,
		PackageName: strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", ""),
		LibraryURL:  libraryURL,
		ModulePath:  modulePath,
	})
}

func getServerLibraryURL(serverVersion string) (url string, tag string, err error) {
	if serverVersion == "latest" {
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
