package run

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Bananenpro/cli"
	cgExec "github.com/code-game-project/codegame-cli/pkg/exec"
)

func RunClient(url string, args ...string) error {
	cmdArgs := []string{"run", "."}
	cmdArgs = append(cmdArgs, args...)

	env := []string{"CG_GAME_URL=" + url}
	env = append(env, os.Environ()...)

	if _, err := exec.LookPath("go"); err != nil {
		cli.Error("'go' ist not installed!")
		return err
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		cli.Error("Failed to run 'CG_GAME_URL=%s go %s'", url, strings.Join(cmdArgs, " "))
	}
	return nil
}

func RunServer(args ...string) error {
	cmdArgs := []string{"run", "."}
	cmdArgs = append(cmdArgs, args...)
	_, err := cgExec.Execute(false, "go", cmdArgs...)
	return err
}
