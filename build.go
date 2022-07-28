package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/code-game-project/codegame-cli/pkg/cgfile"
	cgExec "github.com/code-game-project/codegame-cli/pkg/exec"
	"github.com/code-game-project/codegame-cli/pkg/modules"
)

func Build() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.BuildData]()
	if err != nil {
		return err
	}

	switch config.Type {
	case "client":
		return buildClient(config.Game, data.Output, config.URL)
	case "server":
		return buildServer(data.Output)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func buildClient(gameName, output, url string) error {
	out, err := getOutputName(output, false)
	if err != nil {
		return err
	}
	packageName, err := GetGoModuleName("")
	if err != nil {
		return err
	}
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	cmdArgs := []string{"build", "-o", out, "-ldflags", fmt.Sprintf("-X %s/%s.URL=%s", packageName, gamePackageName, url)}

	_, err = cgExec.Execute(false, "go", cmdArgs...)
	return err
}

func buildServer(output string) error {
	out, err := getOutputName(output, true)
	if err != nil {
		return err
	}
	cmdArgs := []string{"build", "-o", out}
	_, err = cgExec.Execute(false, "go", cmdArgs...)
	return err
}

func getOutputName(output string, isServer bool) (string, error) {
	absRoot, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if output == "" {
		output = filepath.Base(absRoot)
		if isServer {
			output += "-server"
		}
	}

	if runtime.GOOS == "windows" && !strings.HasSuffix(output, ".exe") {
		output += ".exe"
	}

	if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		return "", fmt.Errorf("'%s' already exists and is a directory. Specify another output name with '-o <name>'.", output)
	}

	return output, nil
}
