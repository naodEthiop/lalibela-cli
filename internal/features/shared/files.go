package shared

import (
	"os"
	"path/filepath"
)

// WriteFileIfMissing writes content to a project file if the file does not
// already exist.
func WriteFileIfMissing(projectRoot, relativePath string, content []byte) error {
	fullPath := filepath.Join(projectRoot, relativePath)
	if _, err := os.Stat(fullPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, content, 0o644)
}
