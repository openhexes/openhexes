package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func LocateAppRoot() (string, error) {
	const (
		maxSearchLevel = 5
		markerDir      = ".devcontainer"
	)

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	dir := wd
	for range maxSearchLevel {
		content, err := os.ReadDir(dir)
		if err != nil {
			return "", fmt.Errorf("reading directory: %q: %w", dir, err)
		}
		for _, d := range content {
			if d.IsDir() && d.Name() == markerDir {
				return dir, nil
			}
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf(
		"app root is too far (>%d) from working directory: %q (searching for %q)",
		maxSearchLevel, wd, markerDir,
	)
}
