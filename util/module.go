package util

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
)

func GetModuleName(projectRoot string) (string, error) {
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
