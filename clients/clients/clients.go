package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
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
func (c *ClientSession) callEndpoint(method, url string, body string) (string, error) {
	client := http.Client{}
	var request *http.Request
	var payload io.Reader
	if len(body) != 0 {
		payload = bytes.NewReader([]byte(body))
	}

	if req, err := http.NewRequest(method, url, payload); err != nil {
		return "", err
	} else {
		request = req
	}

	request.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {c.authentication},
		"session-id":    {c.sessionId},
	}

	if resp, err := client.Do(request); err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return "", fmt.Errorf("invalid call: %s", resp.Status)
		} else if body, err := io.ReadAll(resp.Body); err != nil {
			return "", err
		} else {
			return string(body), err
		}
	}
}

// GetUsername access server at that endpoint and returns this content
func (c *ClientSession) GetUsername() (string, error) {
	return c.callEndpoint("GET", CONNECTION_BASE+"user/whoami", "")
}

// SetUserPassword changes user password
func (c *ClientSession) SetUserPassword(password string) error {
	_, err := c.callEndpoint("POST", CONNECTION_BASE+"user/password", password)
	return err
}

// AddUser is an admin task that adds an user with that username and password.
// NOTE THAT this user has no role at all once created
func (c *ClientSession) AddUser(username, password string) error {
	payload := map[string]string{"name": username, "password": password}
	if body, err := json.Marshal(payload); err != nil {
		return err
	} else if _, err := c.callEndpoint("POST", CONNECTION_BASE+"admin/user/create", string(body)); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes user by login
func (c *ClientSession) DeleteUser(username string) error {
	_, err := c.callEndpoint("DELETE", CONNECTION_BASE+"root/user/delete/"+username, "")
	return err
}

// GetUserRoles gets the resources and linked roles for current user (needs admin or root)
func (c *ClientSession) GetUserRoles(username string) (map[string][]string, error) {
	result := make(map[string][]string)
	var loaded map[string]any
	if payload, err := c.callEndpoint("GET", CONNECTION_BASE+"admin/user/roles/"+username, ""); err != nil {
		return nil, err
	} else if err := json.Unmarshal([]byte(payload), &loaded); err != nil {
		return nil, err
	} else {
		for group, v := range loaded {
			if values, ok := v.([]any); !ok {
				return nil, fmt.Errorf("malformed input")
			} else {
				var roles []string
				for _, value := range values {
					if role, ok := value.(string); !ok {
						return nil, fmt.Errorf("malformed input")
					} else {
						roles = append(roles, role)
					}
				}

				result[group] = roles
			}
		}
	}

	return result, nil
}
