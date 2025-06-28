package engines

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"slices"
)

// RequestDecorator returns a decorator over a request
type RequestDecorator struct {
	content    []byte
	bodyClosed bool
	request    *http.Request
}

// NewRequestDecorator returns a new decorated request
func NewRequestDecorator(r *http.Request) RequestDecorator {
	return RequestDecorator{request: r}
}

// GetPath returns the path of the request
func (r *RequestDecorator) GetPath() string {
	return r.request.URL.Path
}

// Header returns request header
func (r *RequestDecorator) Header() http.Header {
	return r.request.Header
}

// HeaderByNames builds header from original source by picking only the ones with given names
func (r *RequestDecorator) HeaderByNames(names ...string) http.Header {
	result := make(http.Header)
	if len(names) == 0 {
		return result
	}

	for k, v := range r.request.Header {
		if slices.Contains(names, k) {
			cloned := make([]string, len(v))
			copy(cloned, v)
			result[k] = cloned
		}
	}

	return result
}

// loadBody is an internal function that loads the request body if not set yet
func (r *RequestDecorator) loadBody() error {
	if !r.bodyClosed {
		defer r.request.Body.Close()
		r.bodyClosed = true
		if body, err := io.ReadAll(r.request.Body); err != nil {
			return err
		} else {
			r.content = body
		}
	}

	return nil
}

// GetRequestMethod returns the request method (PUT, GET, etc)
func (r *RequestDecorator) GetRequestMethod() string {
	return r.request.Method
}

// GetQueryParameters returns the query parameters
// For instance /test/{param} applied to /test/value would return param => value
func (r *RequestDecorator) GetQueryParameters() map[string]string {
	result := make(map[string]string)
	pattern := r.request.Pattern
	parameters := regexp.MustCompile(`\{[a-zA-Z]+\}`).FindAllString(pattern, -1)
	for _, parameter := range parameters {
		parameter = parameter[1 : len(parameter)-1]
		value := r.request.PathValue(parameter)
		result[parameter] = value
	}

	return result
}

// GetUrlParameters returns URL parameters as a map of key, values.
// For instance: test?a=b will return a => [b]
func (r *RequestDecorator) GetUrlParameters() map[string][]string {
	result := make(map[string][]string)
	for k, v := range r.request.URL.Query() {
		values := make([]string, len(v))
		copy(values, v)
		result[k] = values
	}

	return result
}

// GetBodyAsString returns the request body as a string
func (r *RequestDecorator) GetBodyAsString() (string, error) {
	var content string
	if err := r.loadBody(); err != nil {
		return content, err
	} else {
		content = string(r.content)
		return content, err
	}
}

// GetHeaderFirstValue returns, if any, the first value for that key in the header
func (r *RequestDecorator) GetHeaderFirstValue(key string) string {
	var result string
	if values, found := r.request.Header[key]; found && len(values) != 0 {
		result = values[0]
	}

	return result
}

// BindJsonBody fills v from the body assumed to be json
func (r *RequestDecorator) BindJsonBody(v any) error {
	if err := r.loadBody(); err != nil {
		return err
	}

	return json.Unmarshal(r.content, v)
}
