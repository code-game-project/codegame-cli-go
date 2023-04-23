package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-game-project/cli-module/module"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/versions"
)

var FeedbackPkg = feedback.Package("module-go")

// populated by CI
var version = "0.0"

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	projectName := filepath.Base(workingDir)

	module.Run("go", "Go", versions.MustParse(version), map[modules.ProjectType][]versions.Version{
		modules.ProjectType_CLIENT: {versions.MustParse("0.9")},
		modules.ProjectType_SERVER: {versions.MustParse("0.9")},
	}, module.Config{
		Create:    create(projectName),
		RunClient: runClient,
		RunServer: runServer,
	}, feedback.SeverityInfo)
}
