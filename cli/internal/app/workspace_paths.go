package app

import (
	"os"
	"path/filepath"
)

func defaultProjectPath(name string) string {
	// Running at repo root.
	if dirExists("cli") && dirExists("www") {
		return name
	}

	// Running from cli/ (for example via `make cli-run`).
	if dirExists(filepath.Join("..", "cli")) && dirExists(filepath.Join("..", "www")) {
		return filepath.Join("..", name)
	}

	return name
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
