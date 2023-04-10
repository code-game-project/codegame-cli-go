package main

import (
	"errors"
	"fmt"

	"github.com/code-game-project/cli-utils/modules"
)

func Create(projectName string) error {
	actionData := modules.GetCreateData()
	fmt.Println(actionData)
	switch actionData.ProjectType {
	case modules.ProjectType_CLIENT:
		return CreateClient()
	case modules.ProjectType_SERVER:
		return CreateServer(projectName)
	}
	return errors.New("unknown project type")
}
