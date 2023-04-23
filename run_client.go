package main

import (
	"fmt"
	"os"
	"os/exec"

	cgexec "github.com/code-game-project/cli-utils/exec"

	"github.com/code-game-project/cli-utils/modules"
)

func runClient(data *modules.ActionRunClientData) error {
	cmdArgs := []string{"run", "."}
	cmdArgs = append(cmdArgs, data.Args...)
	env := os.Environ()
	env = append(env, []string{"CG_GAME_URL=" + data.GameURL, "CG_GAME_ID=" + data.GameID}...)
	if data.Spectate {
		env = append(env, "CG_SPECTATE=1")
	} else {
		env = append(env, []string{"CG_PLAYER_ID=" + data.GetPlayerID(), "CG_PLAYER_SECRET=" + data.GetPlayerSecret()}...)
	}

	if !cgexec.IsInstalled("go") {
		return fmt.Errorf("'go' is not installed")
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	return cmd.Run()
}
