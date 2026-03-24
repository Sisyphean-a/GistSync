package pathmap

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const homePlaceholder = "{{HOME}}"

var errEmptyPath = errors.New("path cannot be empty")

func ExpandHomePath(path string) (string, error) {
	if path == "" {
		return "", errEmptyPath
	}
	if !strings.Contains(path, homePlaceholder) {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	replaced := strings.Replace(path, homePlaceholder, homeDir, 1)
	return filepath.Clean(replaced), nil
}
