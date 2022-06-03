package server

import (
	"fmt"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/code-game-project/codegame-cli-go/util"
	"github.com/code-game-project/codegame-cli/cli"
	"github.com/code-game-project/codegame-cli/external"
)

//go:embed templates/main.go.tmpl
var goServerMainTemplate string

//go:embed templates/game.go.tmpl
var goServerGameTemplate string

//go:embed templates/events.cge.tmpl
var goServerCGETemplate string

//go:embed templates/event_definitions.go.tmpl
var goServerEventsTemplate string

func CreateNewServer(projectName string) error {
	module, err := cli.Input("Project module path:")
	if err != nil {
		return err
	}

	out, err := external.ExecuteHidden("go", "mod", "init", module)
	if err != nil {
		if out != "" {
			cli.Error(out)
		}
		return err
	}

	cli.Begin("Fetching latest version numbers...")
	cgeVersion, err := external.LatestCGEVersion()
	if err != nil {
		return err
	}

	libraryURL, err := getServerLibraryURL()
	if err != nil {
		return err
	}
	cli.Finish()

	cli.Begin("Creating project template...")
	err = createGoServerTemplate(projectName, module, cgeVersion, libraryURL)
	if err != nil {
		return err
	}
	cli.Finish()

	cli.Begin("Installing dependencies...")

	out, err = external.ExecuteHidden("go", "mod", "tidy")
	if err != nil {
		if out != "" {
			cli.Error(out)
		}
		return err
	}

	cli.Finish()

	cli.Begin("Organizing imports...")

	if !external.IsInstalled("goimports") {
		cli.Warn("Failed to organize import statements: 'goimports' is not installed!")
		return nil
	}
	external.ExecuteHidden("goimports", "-w", "main.go")

	packageDir := strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", "")
	external.ExecuteHidden("goimports", "-w", filepath.Join(packageDir, "game.go"))

	cli.Finish()

	return nil
}

func createGoServerTemplate(projectName, module, cgeVersion, libraryURL string) error {
	err := executeGoServerTemplate(goServerMainTemplate, "main.go", projectName, cgeVersion, libraryURL, module)
	if err != nil {
		return err
	}

	err = executeGoServerTemplate(goServerCGETemplate, "events.cge", projectName, cgeVersion, libraryURL, module)
	if err != nil {
		return err
	}

	packageName := strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", "")

	err = executeGoServerTemplate(goServerGameTemplate, filepath.Join(packageName, "game.go"), projectName, cgeVersion, libraryURL, module)
	if err != nil {
		return err
	}

	return executeGoServerTemplate(goServerEventsTemplate, filepath.Join(packageName, "event_definitions.go"), projectName, cgeVersion, libraryURL, module)
}

func executeGoServerTemplate(templateText, fileName, projectName, cgeVersion, libraryURL, modulePath string) error {
	type data struct {
		Name          string
		PackageName   string
		SnakeCaseName string
		CGEVersion    string
		LibraryURL    string
		ModulePath    string
	}

	return util.ExecTemplate(templateText, fileName, data{
		Name:          projectName,
		PackageName:   strings.ReplaceAll(strings.ReplaceAll(projectName, "_", ""), "-", ""),
		SnakeCaseName: strings.ReplaceAll(projectName, "-", "_"),
		CGEVersion:    cgeVersion,
		LibraryURL:    libraryURL,
		ModulePath:    modulePath,
	})
}

func getServerLibraryURL() (string, error) {
	tag, err := external.LatestGithubTag("code-game-project", "go-server")
	if err != nil {
		return "", err
	}
	majorVersion := strings.TrimPrefix(strings.Split(tag, ".")[0], "v")

	path := "github.com/code-game-project/go-server/cg"
	if majorVersion != "0" && majorVersion != "1" {
		path = fmt.Sprintf("github.com/code-game-project/go-server/v%s/cg", majorVersion)
	}
	return path, nil
}
