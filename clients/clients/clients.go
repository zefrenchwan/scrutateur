package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const CONNECTION_BASE = "http://localhost:3000/"

// ClientSession has the main info to use the application: session id and authentication token
type ClientSession struct {
	sessionId      string
	authentication string
}

// Connect validates user auth info and sets the context (auth info) for the rest of the calls
func Connect(login, password string) (ClientSession, error) {
	var result ClientSession
	payload, errMarshal := json.Marshal(map[string]string{"login": login, "password": password})
	if errMarshal != nil {
		panic(errMarshal)
	}

	if resp, err := http.Post(CONNECTION_BASE+"login", "application/json", bytes.NewReader(payload)); err != nil {
		return result, err
	} else {
		result.sessionId = resp.Header.Get("session-id")
		result.authentication = resp.Header.Get("Authorization")
		return result, nil
	}
}

// callEndpoint is the low level http call mechanism
func (c *ClientSession) callEndpoint(method, url string) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {c.authentication},
		"session-id":    {c.sessionId},
	}

	if resp, err := client.Do(req); err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err != nil {
			return "", err
		} else {
			return string(body), err
		}
	}
}

// GetUserDetails access server at that endpoint and returns this content
func (c *ClientSession) GetUserDetails() (string, error) {
	return c.callEndpoint("GET", CONNECTION_BASE+"user/details")
}
