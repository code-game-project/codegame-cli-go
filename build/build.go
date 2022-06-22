package build

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Bananenpro/cli"
	cgExec "github.com/code-game-project/codegame-cli/util/exec"
)

func BuildClient(projectRoot, gameName, output, url string) error {
	out, err := getOutputName(projectRoot, output)
	if err != nil {
		return err
	}
	packageName, err := getPackageName(projectRoot)
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

func getPackageName(projectRoot string) (string, error) {
	path := filepath.Join(projectRoot, "go.mod")
	file, err := os.Open(path)
	if err != nil {
		return "", cli.Error("Failed to open '%s'", path)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "module ") {
			return strings.TrimSuffix(strings.TrimPrefix(scanner.Text(), "module "), "/"), nil
		}
	}
	if scanner.Err() != nil {
		cli.Error("Failed to read '%s': %s", path, scanner.Err())
	}
	return "", cli.Error("Missing 'module' statement in '%s'", path)
}
