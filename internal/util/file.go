package util

import (
	"fmt"
	"os"
	"path/filepath"
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
