package clients

import (
	"fmt"
	"io"
	"net/http"
)

// STATIC_FILE_PREFIX is the prefix for static resources to load
const STATIC_FILE_PREFIX = CONNECTION_BASE + "app/static/"

// LoadStaticContent tries to load a relative resource (for instance changelog.txt)
func LoadStaticContent(path string) ([]byte, error) {
	client := http.Client{}
	var request *http.Request

	if req, err := http.NewRequest("GET", STATIC_FILE_PREFIX+path, nil); err != nil {
		return nil, err
	} else {
		request = req
	}

	if resp, err := client.Do(request); err != nil {
		return nil, err
	} else {
		// read content
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("invalid call: %s", resp.Status)
		} else if body, err := io.ReadAll(resp.Body); err != nil {
			return nil, err
		} else {
			return body, nil
		}
	}
}
