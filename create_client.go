package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "embed"

	"github.com/code-game-project/cge-parser/adapter"
	"github.com/code-game-project/cli-module/module"
	"github.com/code-game-project/cli-utils/casing"
	"github.com/code-game-project/cli-utils/cli"
	"github.com/code-game-project/cli-utils/exec"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/request"
	"github.com/code-game-project/cli-utils/templates"
	"github.com/code-game-project/cli-utils/versions"
	gomod "golang.org/x/mod/module"
)

//go:embed templates/new/client/main.go.tmpl
var clientMainTemplate string

//go:embed templates/new/client/game.go.tmpl
var clientGameTemplate string

//go:embed templates/new/client/events.go.tmpl
var clientEventsTemplate string

//go:embed templates/new/gitignore.tmpl
var gitignoreTemplate string

var simpleModPathRegex = regexp.MustCompile("[a-z]+[a-z0-9]*")

func CreateClient(data *modules.ActionCreateData, projectName string) error {
	modulePath := cli.Input("Project module path:", true, casing.ToOneWord(projectName), func(input interface{}) error {
		if simpleModPathRegex.Match([]byte(input.(string))) {
			return nil
		}
		err := gomod.CheckPath(input.(string))
		if err != nil {
			return fmt.Errorf("invalid module path: %w", err.(*gomod.InvalidPathError).Err)
		}
		return nil
	})

	feedback.Info(FeedbackPkg, "Initializing Go module...")
	err := exec.ExecuteDimmed("go", "mod", "init", modulePath)
	if err != nil {
		return fmt.Errorf("create go module: %w", err)
	}

	libraryURL, libraryTag, err := getClientLibraryURL(data.LibraryVersion)
	if err != nil {
		return fmt.Errorf("determine client library details: %w", err)
	}

	feedback.Info(FeedbackPkg, "Installing go-client %s...", libraryTag)
	err = exec.ExecuteDimmed("go", "get", fmt.Sprintf("%s@%s", libraryURL, libraryTag))
	if err != nil {
		return fmt.Errorf("install go-client: %w", err)
	}

	feedback.Info(FeedbackPkg, "Loading CGE data...")
	cgeData, err := module.LoadCGEData(data.GetUrl())
	if err != nil {
		return fmt.Errorf("load cge data: %w", err)
	}

	feedback.Info(FeedbackPkg, "Generating template...")
	err = createClientTemplate(modulePath, data.GameName, libraryURL, cgeData)
	if err != nil {
		return err
	}

	feedback.Info(FeedbackPkg, "Installing dependencies...")
	err = exec.ExecuteDimmed("go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("install dependencies: %w", err)
	}

	return nil
}

func createClientTemplate(modulePath, gameName, libraryURL string, cgeData adapter.ParserResponse) error {
	return execClientTemplate(modulePath, gameName, libraryURL, cgeData, false)
}

func execClientTemplate(modulePath, gameName, libraryURL string, cgeData adapter.ParserResponse, update bool) error {
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")
	gameDir := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	if update {
		feedback.Warn(FeedbackPkg, "This action will ERASE and regenerate ALL files in '%s/'.\nYou will have to manually update your code to work with the new version.", gameDir)
		ok := cli.YesNo("Continue?", false)
		if !ok {
			os.Exit(2)
		}
		os.RemoveAll(gameDir)
	} else {
		feedback.Warn(FeedbackPkg, "DO NOT EDIT the `%s/` directory inside of the project. ALL CHANGES WILL BE LOST when running `codegame update`.", gameDir)
	}

	type event struct {
		Name       string
		PascalName string
	}

	eventNames := make([]event, len(cgeData.Events))
	for i, e := range cgeData.Events {
		pascal := strings.ReplaceAll(e.Name, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		eventNames[i] = event{
			Name:       e.Name,
			PascalName: pascal,
		}
	}

	commandNames := make([]event, len(cgeData.Commands))
	for i, c := range cgeData.Commands {
		pascal := strings.ReplaceAll(c.Name, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		commandNames[i] = event{
			Name:       c.Name,
			PascalName: pascal,
		}
	}

	data := struct {
		LibraryURL  string
		PackageName string
		ModulePath  string
		Events      []event
		Commands    []event
	}{
		LibraryURL:  libraryURL,
		PackageName: gamePackageName,
		ModulePath:  modulePath,
		Events:      eventNames,
		Commands:    commandNames,
	}

	if !update {
		err := templates.Execute(clientMainTemplate, "main.go", data)
		if err != nil {
			return err
		}

		err = templates.Execute(gitignoreTemplate, ".gitignore", data)
		if err != nil {
			return err
		}
	}

	err := templates.Execute(clientGameTemplate, filepath.Join(gameDir, "game.go"), data)
	if err != nil {
		return err
	}

	err = templates.Execute(clientEventsTemplate, filepath.Join(gameDir, "events.go"), data)
	if err != nil {
		return err
	}

	eventDefinitionsFile, err := os.Create(filepath.Join(gameDir, "event_definitions.go"))
	if err != nil {
		return fmt.Errorf("create event_definitions.go: %w", err)
	}
	generateEventDefinitions(eventDefinitionsFile, gamePackageName, libraryURL, cgeData)
	eventDefinitionsFile.Close()
	return nil
}

func getClientLibraryURL(libVersion *string) (url string, tag string, err error) {
	type tagsResp []struct {
		Name string `json:"name"`
	}
	tags, err := request.FetchJSON[tagsResp]("https://api.github.com/repos/code-game-project/go-client/tags", 24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("fetch go-client git tags: %w", err)
	}
	if len(tags) == 0 {
		return "", "", versions.ErrNoCompatibleVersion
	}
	if libVersion == nil {
		temp := strings.TrimPrefix(strings.Join(strings.Split(tags[0].Name, ".")[:2], "."), "v")
		libVersion = &temp
	}
	version, err := versions.Parse(*libVersion)
	if err != nil {
		return "", "", fmt.Errorf("invalid client library version: %w", err)
	}

	for _, t := range tags {
		if strings.HasPrefix(t.Name, "v"+(*libVersion)) {
			tag = t.Name
			break
		}
	}
	if tag == "" {
		return "", "", versions.ErrNoCompatibleVersion
	}

	path := "github.com/code-game-project/go-client/cg"
	if version[0] > 1 {
		path = fmt.Sprintf("github.com/code-game-project/go-client/v%d/cg", version[0])
	}

	return path, tag, nil
}
