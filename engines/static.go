package engines

import (
	"net/http"
	"strings"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

// ListStaticResources maps a directory content to a set of resources.
// For instance, assume that urlBase is /resources/
// when reading /app/static/ folder on a local FS, we find img.png
// thenr result would be /resources/img.png => /app/static/img.png
func ListStaticResources(urlBase string, localBase string) (map[string]string, error) {
	if paths, err := storage.ListDirectoryContent(localBase); err != nil {
		return nil, err
	} else {
		// make parameters as standard input
		urlPrefix := urlBase
		if !strings.HasSuffix(urlPrefix, "/") {
			urlPrefix = urlPrefix + "/"
		}
		localPrefix := localBase
		if !strings.HasSuffix(localPrefix, "/") {
			localPrefix = localPrefix + "/"
		}

		// basically map the fs path to the url accessible resource
		result := make(map[string]string)
		for _, path := range paths {
			key := strings.Replace(path, localPrefix, urlPrefix, 1)
			result[key] = path
		}

		return result, nil
	}
}

// BuildStaticHandlerForLocalResource builds a request processor to get a given resource
func BuildStaticHandlerForLocalResource(mappedUrl, localPath string) RequestProcessor {
	return func(context *HandlerContext) error {
		url := context.GetRequestPath()
		if mappedUrl == url {
			if body, err := storage.LoadContentFromFS(localPath); err != nil {
				context.BuildError(http.StatusInternalServerError, err, nil)
			} else {
				context.BuildRaw(http.StatusOK, body, context.RequestHeaderByNames("Authorization"))
			}
		} else {
			context.Build(http.StatusNotFound, "", nil)
		}

		return nil
	}
}
