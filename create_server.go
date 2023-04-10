package main

import (
	"fmt"
	"path/filepath"
	"strings"

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

func CreateServer(projectName string) error {
	fmt.Println("create server")
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
