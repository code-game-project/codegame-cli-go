package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-go/util"
	cgExec "github.com/code-game-project/codegame-cli/util/exec"
)

func BuildClient(projectRoot, gameName, output, url string) error {
	out, err := getOutputName(projectRoot, output)
	if err != nil {
		return err
	}
	packageName, err := util.GetModuleName(projectRoot)
	if err != nil {
		return err
	}
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	cmdArgs := []string{"build", "-o", out, "-ldflags", fmt.Sprintf("-X %s/%s.URL=%s", packageName, gamePackageName, url), filepath.Join(projectRoot, "main.go")}

	_, err = cgExec.Execute(false, "go", cmdArgs...)
	return err
}

func BuildServer(projectRoot, output string) error {
	out, err := getOutputName(projectRoot, output)
	if err != nil {
		return err
	}
	cmdArgs := []string{"build", "-o", out, filepath.Join(projectRoot, "main.go")}
	_, err = cgExec.Execute(false, "go", cmdArgs...)
	return err
}

func getOutputName(projectRoot, output string) (string, error) {
	absRoot, err := filepath.Abs(projectRoot)
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
