package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func MakeDirs(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

func ExpandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	// Get the home directory from environment variable
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Replace the tilde (~) with the actual path
	return filepath.Join(homeDir, path[2:]), nil
}
