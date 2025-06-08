package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ListDirectoryContent gets all the resources within a dir and returns their absolute path
func ListDirectoryContent(basePath string) ([]string, error) {
	var dirPath string
	if strings.HasSuffix(basePath, "/") {
		dirPath = basePath
	} else {
		dirPath = basePath + "/"
	}

	// all paths
	var content []string
	// global error
	var resultingError error
	// Using WalkDir and collecting
	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			resultingError = errors.Join(resultingError, err)
		} else if !d.IsDir() {
			content = append(content, path)
		}

		return err
	})

	return content, resultingError
}

// LoadContentFromFS gets the content of a file
func LoadContentFromFS(path string) ([]byte, error) {
	return os.ReadFile(path)
}
