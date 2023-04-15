package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/code-game-project/cli-utils/casing"
	"github.com/code-game-project/cli-utils/cli"
	"github.com/code-game-project/cli-utils/exec"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/request"
	"github.com/code-game-project/cli-utils/versions"

	gomod "golang.org/x/mod/module"
)

func create(projectName string) func(data *modules.ActionCreateData) error {
	return func(data *modules.ActionCreateData) error {
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
		switch data.ProjectType {
		case modules.ProjectType_CLIENT:
			return CreateClient(data, modulePath)
		case modules.ProjectType_SERVER:
			return CreateServer(data, projectName, modulePath)
		}
		return nil
	}
}

func getLibraryURL(libraryName string, libVersion *string) (url string, tag string, err error) {
	type tagsResp []struct {
		Name string `json:"name"`
	}
	tags, err := request.FetchJSON[tagsResp](fmt.Sprintf("https://api.github.com/repos/code-game-project/%s/tags", libraryName), 24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("fetch %s git tags: %w", libraryName, err)
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
		return "", "", fmt.Errorf("invalid %s version: %w", libraryName, err)
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

	path := fmt.Sprintf("github.com/code-game-project/%s/cg", libraryName)
	if version[0] > 1 {
		path = fmt.Sprintf("github.com/code-game-project/%s/v%d/cg", libraryName, version[0])
	}

	return path, tag, nil
}
