package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	cgExec "github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
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
	data.OS = strings.ReplaceAll(data.OS, "current", "")
	data.OS = strings.ReplaceAll(data.OS, "macos", "darwin")
	data.Arch = strings.ReplaceAll(data.Arch, "current", "")
	data.Arch = strings.ReplaceAll(data.Arch, "x86", "386")
	data.Arch = strings.ReplaceAll(data.Arch, "x64", "amd64")
	data.Arch = strings.ReplaceAll(data.Arch, "arm32", "arm")

	switch config.Type {
	case "client":
		return buildClient(config.Game, data.Output, config.URL, data.OS, data.Arch)
	case "server":
		return buildServer(data.Output, data.OS, data.Arch)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func buildClient(gameName, output, url, operatingSystem, architecture string) error {
	cli.BeginLoading("Building...")
	out, err := getOutputName(output, false, operatingSystem)
	if err != nil {
		return err
	}
	packageName, err := GetGoModuleName("")
	if err != nil {
		return err
	}
	gamePackageName := strings.ReplaceAll(strings.ReplaceAll(gameName, "-", ""), "_", "")

	cmdArgs := []string{"build", "-o", out, "-ldflags", fmt.Sprintf("-X %s/%s.URL=%s", packageName, gamePackageName, url)}

	if _, err = exec.LookPath("go"); err != nil {
		return fmt.Errorf("'go' ist not installed!")
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Env = append(os.Environ(), "GOOS="+operatingSystem, "GOARCH="+architecture)

	buildOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(buildOutput))
		return fmt.Errorf("Failed to run 'GOOS=%s GOARCH=%s go %s'", operatingSystem, architecture, strings.Join(cmdArgs, " "))
	}
	cli.FinishLoading()
	return nil
}

func buildServer(output, operatingSystem, architecture string) error {
	cli.BeginLoading("Building...")
	out, err := getOutputName(output, true, operatingSystem)
	if err != nil {
		return err
	}
	cmdArgs := []string{"build", "-o", out}
	_, err = cgExec.Execute(false, "go", cmdArgs...)

	if _, err = exec.LookPath("go"); err != nil {
		return fmt.Errorf("'go' ist not installed!")
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Env = append(os.Environ(), "GOOS="+operatingSystem, "GOARCH="+architecture)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run 'GOOS=%s GOARCH=%s go %s'", operatingSystem, architecture, strings.Join(cmdArgs, " "))
	}
	cli.FinishLoading()
	return nil
}

func getOutputName(output string, isServer bool, operatingSystem string) (string, error) {
	absRoot, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if output == "" {
		output = filepath.Base(absRoot)
		if isServer {
			if stat, err := os.Stat(output); err == nil && stat.IsDir() {
				output += "-server"
			}
		}
	}

	if ((operatingSystem == "" && runtime.GOOS == "windows") || operatingSystem == "windows") && !strings.HasSuffix(output, ".exe") {
		output += ".exe"
	}

	if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		return "", fmt.Errorf("'%s' already exists and is a directory. Specify another output name with '-o <name>'.", output)
	}

	return output, nil
}
