package engines

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
)

// BuildStaticHandler maps a url base to a local storage and gets content from it
func BuildStaticHandler(urlBase string, localBase string) RequestProcessor {
	localProcessor := NewLocalResourceLoader(urlBase, localBase)
	return func(context *HandlerContext) error {
		url := context.request.GetPath()
		if !localProcessor.Accept(url) {
			context.Build(http.StatusNotFound, fmt.Sprintf("no file matching url %s", url), nil)
		} else if body, err := localProcessor.Load(url); err != nil {
			context.BuildError(http.StatusInternalServerError, err, nil)
		} else {
			context.BuildRaw(http.StatusOK, body, context.request.HeaderByNames("Authorization"))
		}

		return nil
	}
}

// LoadStaticResources gets all the resources within a dir and returns their absolute path
func LoadStaticResources(basePath string) ([]string, error) {
	// all paths
	var content []string
	// global error
	var resultingError error
	// Using WalkDir and collecting
	filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			resultingError = errors.Join(resultingError, err)
		} else if !d.IsDir() {
			content = append(content, path)
		}

		return err
	})

	return content, resultingError
}
