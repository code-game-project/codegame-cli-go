package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/util"
	cgExec "github.com/code-game-project/codegame-cli/pkg/exec"
	"github.com/code-game-project/codegame-cli/pkg/modules"
)

func BuildClient(gameName, url string) error {
	data, err := modules.ReadCommandConfig[modules.BuildData]()
	if err != nil {
		return err
	}
	out, err := getOutputName(data.Output)
	if err != nil {
		return err
	}
	packageName, err := util.GetModuleName()
	if err != nil {
		return err
	}
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	cmdArgs := []string{"build", "-o", out, "-ldflags", fmt.Sprintf("-X %s/%s.URL=%s", packageName, gamePackageName, url)}

	_, err = cgExec.Execute(false, "go", cmdArgs...)
	return err
}

func BuildServer() error {
	data, err := modules.ReadCommandConfig[modules.BuildData]()
	if err != nil {
		return err
	}
	out, err := getOutputName(data.Output)
	if err != nil {
		return err
	}
	cmdArgs := []string{"build", "-o", out}
	_, err = cgExec.Execute(false, "go", cmdArgs...)
	return err
}

func getOutputName(output string) (string, error) {
	absRoot, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if output == "" {
		output = filepath.Base(absRoot)
	}

	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		return "", cli.Error("'%s' already exists and is a directory. Specify another output name with '-o <name>'.", output)
	}

	return output, nil
}
