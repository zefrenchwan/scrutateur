package engines

import (
	"context"
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/dto"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

// HandlerContextKey is necessary as a key type for context handling
type HandlerContextKey string

// ProcessingAuth defines the auth content for now
type ProcessingAuth struct {
	Login string
	Roles []dto.GrantRole
}

// HandlerContext is the context to pass on each request, for the processor to get everything
type HandlerContext struct {
	// go context
	context context.Context
	// decorated request (to ease request use)
	request RequestDecorator
	// response to send, will do as the last step
	response AbstractResponse
	// design choice: include dao in here, we don't build a general engine, we want custom use
	Dao storage.Dao
	// current auth content as structured data (no "any" stuff)
	CurrentAuth ProcessingAuth
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

// RequestUrlParameters returns the URL parameters as a map of string and related values
func (c *HandlerContext) RequestUrlParameters() map[string][]string {
	return c.request.GetUrlParameters()
}

// IsAuthDefined returns true if login is set and we know who's asking
func (c *HandlerContext) IsAuthDefined() bool {
	return c.CurrentAuth.Login != ""
}

// SetLogin registers this user login for that request
func (c *HandlerContext) SetLogin(value string) {
	c.CurrentAuth.Login = value
}

// GetLogin returns the current login
func (c *HandlerContext) GetLogin() string {
	return c.CurrentAuth.Login
}

// GetRoles returns the current roles for that query to process
func (c *HandlerContext) GetRoles() []dto.GrantRole {
	return c.CurrentAuth.Roles
}

// SetRoles sets the role for that query to process
func (c *HandlerContext) SetRoles(roles []dto.GrantRole) {
	c.CurrentAuth.Roles = roles
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

// SetRawResponseValues starts a new response with that values
func (ar *HandlerContext) SetRawResponseValues(code int, body []byte, headers http.Header) {
	ar.response.setResponse(code, body, headers)
}

// SetStringResponseValues constructs a new response with a string body, and flags it NOT to be ready
func (ar *HandlerContext) SetStringResponseValues(code int, body string, headers http.Header) {
	ar.response.SetStringResponseValues(code, body, headers)
}

// SetJSONResponseValues constructs a new response with a json serializable body, and flags it NOT to be ready
func (ar *HandlerContext) SetJSONResponseValues(code int, body any, headers http.Header) error {
	return ar.response.SetJSONResponseValues(code, body, headers)
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
