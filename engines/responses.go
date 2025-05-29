package engines

import (
	"encoding/json"
	"errors"
	"net/http"
)

// AbstractResponse is what to send when receiving a query
type AbstractResponse struct {
	// Code defined in http
	Code int
	// Headers contains all headers to send
	Headers http.Header
	// Content will be the body to send
	Content []byte
	// Ready is true when we should send the response ASAP
	Ready bool
}

// End marks current response to be ready to answer
func (ar *AbstractResponse) End() {
	ar.Ready = true
}

// Write writes the content of the response into a http writer (and response is then sent)
func (ar *AbstractResponse) Write(w http.ResponseWriter) error {
	if !ar.ShouldSend() {
		return errors.New("sending partial response")
	}

	// Write headers
	for k, values := range ar.Headers {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	// Write status
	if ar.Code != http.StatusOK {
		w.WriteHeader(ar.Code)
	}

	// Write body
	if ar.Content != nil {
		_, err := w.Write(ar.Content)
		return err
	}

	return nil
}

// SetHeader sets the value for that key no matter previous content
func (ar *AbstractResponse) SetHeader(key, value string) {
	if ar.Headers == nil {
		ar.Headers = make(http.Header)
	}

	ar.Headers.Set(key, value)
}

// ShouldSend returns if the response should be sent now
func (ar *AbstractResponse) ShouldSend() bool {
	return ar != nil && ar.Ready
}

// AppendToBody adds text to current body, at the end
func (ar *AbstractResponse) AppendToBody(values []byte) {
	ar.Content = append(ar.Content, values...)
}

// ClearBody resets the body
func (ar *AbstractResponse) ClearBody() {
	ar.Content = nil
}

// ClearHeaders removes all headers
func (ar *AbstractResponse) ClearHeaders() {
	ar.Headers = make(http.Header)
}

// buildResponse constructs a full response ready to be sent
func (ar *AbstractResponse) buildResponse(code int, content []byte, headers http.Header) {
	ar.Code = code
	if len(content) > 0 {
		ar.Content = content
	} else {
		ar.Content = nil
	}

	// copy headers
	if len(headers) == 0 {
		ar.Headers = make(http.Header)
	} else {
		ar.Headers = headers.Clone()
	}

	// and it is ready
	ar.Ready = true
}

// Build sets all values in parameters and flags the response to be ready
func (ar *AbstractResponse) Build(code int, body string, headers http.Header) {
	ar.buildResponse(code, []byte(body), headers)
}

// BuildJson builds a json as a body, or returns an error for serde issue
func (ar *AbstractResponse) BuildJson(code int, body any, headers http.Header) error {
	if res, err := json.Marshal(body); err != nil {
		return err
	} else {
		ar.buildResponse(code, res, headers)
		return nil
	}
}

// BuildError sets all values in parameters and flags the response to be ready
func (ar *AbstractResponse) BuildError(code int, failure error, headers http.Header) {
	ar.Build(code, failure.Error(), headers)
}
