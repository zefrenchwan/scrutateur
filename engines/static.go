package engines

import (
	"fmt"
	"net/http"
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
