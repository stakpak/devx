package xray

import (
	"os"
	"path/filepath"
	"strings"
)

func scan(configDir string) (map[string]string, error) {
	overlay := map[string]string{}
	err := filepath.Walk(configDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if isFileInScope(path) {
				fileBytes, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				overlay[path] = string(fileBytes)
			}
			return nil
		})
	if err != nil {
		return overlay, err
	}

	return overlay, nil
}

func isFileInScope(path string) bool {
	return strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".ts") ||
		strings.HasSuffix(path, ".py") ||
		strings.HasSuffix(path, ".go") ||
		strings.HasSuffix(path, ".rs") ||
		strings.HasSuffix(path, ".php") ||
		strings.HasSuffix(path, ".rb") ||
		strings.HasSuffix(path, ".cs")
}
