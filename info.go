package main

import (
	"encoding/json"
	"os"

	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/versions"
)

func Info() error {
	return json.NewEncoder(os.Stdout).Encode(modules.ModuleInfo{
		Actions: []modules.Action{modules.ActionInfo, modules.ActionCreate},
		LibraryVersions: map[string][]versions.Version{
			"client": {versions.MustParse("0.6")},
			"server": {versions.MustParse("0.10")},
		},
		ApplicationTypes: []string{"client", "server"},
	})
}
