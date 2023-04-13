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
		return "", fmt.Errorf("open go.mod: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "module ") {
			return strings.TrimSuffix(strings.TrimPrefix(scanner.Text(), "module "), "/"), nil
		}
	}
	if scanner.Err() != nil {
		return "", fmt.Errorf("read go.mod: %w", scanner.Err())
	}
	return "", fmt.Errorf("missing 'module' statement in '%s'", path)
}
