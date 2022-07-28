package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetGoModuleName(projectRoot string) (string, error) {
	path := filepath.Join(projectRoot, "go.mod")
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Failed to open '%s'", path)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "module ") {
			return strings.TrimSuffix(strings.TrimPrefix(scanner.Text(), "module "), "/"), nil
		}
	}
	if scanner.Err() != nil {
		return "", fmt.Errorf("Failed to read '%s': %s", path, scanner.Err())
	}
	return "", fmt.Errorf("Missing 'module' statement in '%s'", path)
}
