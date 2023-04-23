package main

import (
	"fmt"
	"os"
	"os/exec"

	cgexec "github.com/code-game-project/cli-utils/exec"
	"github.com/code-game-project/cli-utils/modules"
)

func runServer(data *modules.ActionRunServerData) error {
	cmdArgs := []string{"run", "."}
	cmdArgs = append(cmdArgs, data.Args...)

	env := os.Environ()
	if data.Port != nil {
		env = append(env, fmt.Sprintf("CG_PORT=%d", data.GetPort()))
	}

	if !cgexec.IsInstalled("go") {
		return fmt.Errorf("'go' is not installed")
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}
	return err
}
