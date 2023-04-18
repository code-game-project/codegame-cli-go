package main

import (
	"fmt"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/code-game-project/cli-utils/exec"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/templates"
)

//go:embed templates/new/server/main.go.tmpl
var serverMainTemplate string

//go:embed templates/new/server/game.go.tmpl
var serverGameTemplate string

//go:embed templates/new/server/definitions.go.tmpl
var serverEventsTemplate string

//go:embed templates/new/server/Dockerfile.tmpl
var serverDockerfileTemplate string

//go:embed templates/new/server/dockerignore.tmpl
var serverDockerignoreTemplate string

func CreateServer(data *modules.ActionCreateData, projectName, modulePath string) error {
	libraryURL, libraryTag, err := getLibraryURL("go-server", data.LibraryVersion)
	if err != nil {
		return fmt.Errorf("determine server library details: %w", err)
	}

	feedback.Info(FeedbackPkg, "Installing go-server %s...", libraryTag)
	err = exec.ExecuteDimmed("go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return fmt.Errorf("install go-server: %w", err)
	}

	feedback.Info(FeedbackPkg, "Generating template...")
	err = createServerTemplate(projectName, modulePath, libraryURL)
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

	return executeServerTemplate(serverEventsTemplate, filepath.Join(packageName, "definitions.go"), projectName, libraryURL, module)
}

func executeServerTemplate(templateText, fileName, projectName, libraryURL, modulePath string) error {
	type data struct {
		Name        string
		PackageName string
		LibraryURL  string
		ModulePath  string
	}

	return templates.Execute(templateText, fileName, data{
		Name:        projectName,
		PackageName: strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", ""),
		LibraryURL:  libraryURL,
		ModulePath:  modulePath,
	})
}
