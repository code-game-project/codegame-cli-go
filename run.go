package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/config"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"

	cgExec "github.com/code-game-project/go-utils/exec"
)

func Run() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.RunData]()
	if err != nil {
		return err
	}

	url := external.TrimURL(config.URL)

	switch config.Type {
	case "client":
		return runClient(url, data.Args)
	case "server":
		return runServer(data.Args)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func runClient(url string, args []string) error {
	cmdArgs := []string{"run", "."}
	cmdArgs = append(cmdArgs, args...)

	env := []string{"CG_GAME_URL=" + url}
	env = append(env, os.Environ()...)

	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("'go' ist not installed!")
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run 'CG_GAME_URL=%s go %s'", url, strings.Join(cmdArgs, " "))
	}
	return nil
}

func runServer(args []string) error {
	cmdArgs := []string{"run", "."}
	cmdArgs = append(cmdArgs, args...)

	conf := config.Load()
	os.Setenv("CG_PORT", strconv.Itoa(conf.DevPort))

	_, err := cgExec.Execute(false, "go", cmdArgs...)
	return err
}
