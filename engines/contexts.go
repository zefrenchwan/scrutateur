package engines

import (
	"context"
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

// HandlerContextKey is necessary as a key type for context handling
type HandlerContextKey string

// HandlerContext is the context to pass on each request, for the processor to get everything
type HandlerContext struct {
	// go context, a map[string]any was not enough due to dao
	context context.Context
	// decorated request (to ease request use)
	request RequestDecorator
	// response to send, will do as the last step
	response AbstractResponse
	// design choice: include dao in here, we don't build a general engine, we want custom use
	Dao storage.Dao
}

// GetCurrentContext returns the current context
func (c *HandlerContext) GetCurrentContext() context.Context {
	return c.context
}

// GetQueryParameters returns the query parameters
func (c *HandlerContext) GetQueryParameters() map[string]string {
	return c.request.GetQueryParameters()
}

// GetRequestPath returns the path of the query
func (c *HandlerContext) GetRequestPath() string {
	return c.request.GetPath()
}

// GetRequestHeader returns the header of the request
func (c *HandlerContext) GetRequestHeader() map[string][]string {
	return c.request.Header()
}

// RequestHeaderByNames builds header from request by picking only the ones with given names
func (c *HandlerContext) RequestHeaderByNames(names ...string) http.Header {
	return c.request.HeaderByNames(names...)
}

// GetRequestHeaderFirstValue returns the first value, if any, with that key
func (c *HandlerContext) GetRequestHeaderFirstValue(key string) string {
	return c.request.GetHeaderFirstValue(key)
}

// RequestBodyAsString returns the request body as a string, or an error when reading
func (c *HandlerContext) RequestBodyAsString() (string, error) {
	return c.request.GetBodyAsString()
}

// SetContextValue sets value within inner context
func (c *HandlerContext) SetContextValue(key HandlerContextKey, value any) {
	if c.context == nil {
		c.context = context.Background()
	}

	c.context = context.WithValue(c.context, key, value)
}

// GetContextValue gets value or nil for no value
func (c *HandlerContext) GetContextValue(key string) any {
	return c.context.Value(key)
}

// GetContextValueAsString gets value or empty for no value or no match
func (c *HandlerContext) GetContextValueAsString(key HandlerContextKey) string {
	value := c.context.Value(key)
	if result, ok := value.(string); ok {
		return result
	} else {
		return ""
	}
}

// SetResponseHeader adds an header with that key and that value
func (c *HandlerContext) SetResponseHeader(key, value string) {
	c.response.SetHeader(key, value)
}

// SetResponseStatus sets the http status
func (c *HandlerContext) SetResponseStatus(code int) {
	c.response.Code = code
}

// GetRequestMethod returns the method (get, post, etc) the request was submitted with
func (c *HandlerContext) GetRequestMethod() string {
	return c.request.GetRequestMethod()
}

// ClearBody resets body
func (c *HandlerContext) ClearBody() {
	c.response.ClearBody()
}

// AddToBody appends value to current body content
func (c *HandlerContext) AddToBody(value []byte) {
	c.response.AppendToBody(value)
}

// BindJsonBody fills v from the body assumed to be json
func (c *HandlerContext) BindJsonBody(v any) error {
	return c.request.BindJsonBody(v)
}

// ClearResponse clear any field linked to response
func (c *HandlerContext) ClearResponse() {
	c.response.ClearBody()
	c.response.ClearHeaders()
	c.response.Code = 0
}

// Build sets all values in parameters and flags the response to be ready
func (c *HandlerContext) Build(code int, body string, headers http.Header) {
	c.response.Build(code, body, headers)
}

// BuildJson builds a json as a body, or returns an error for serde issue
func (c *HandlerContext) BuildJson(code int, body any, headers http.Header) error {
	return c.response.BuildJson(code, body, headers)
}

// BuildError sets all values in parameters and flags the response to be ready
func (c *HandlerContext) BuildError(code int, failure error, headers http.Header) {
	c.response.BuildError(code, failure, headers)
}

// BuildRaw sets values (body as is) and mark the response as ready
func (c *HandlerContext) BuildRaw(code int, body []byte, header http.Header) {
	c.response.buildResponse(code, body, header)
}

// Done flags the response as ready
func (c *HandlerContext) Done() {
	c.response.End()
}
