package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const CONNECTION_BASE = "http://localhost:3000/"
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

// ClientSession has the main info to use the application
type ClientSession struct {
	authorization string
}

// Connect validates user auth info and sets the context (auth info) for the rest of the calls
func Connect(login, password string) (ClientSession, error) {
	var result ClientSession
	payload, errMarshal := json.Marshal(map[string]string{"name": login, "password": password})
	if errMarshal != nil {
		panic(errMarshal)
	}

	if resp, err := http.Post(CONNECTION_BASE+"login", "application/json", bytes.NewReader(payload)); err != nil {
		return result, err
	} else {
		result.authorization = resp.Header.Get("Authorization")
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
		"Authorization": {c.authorization},
	}

	if resp, err := client.Do(request); err != nil {
		// try and see what happened
		defer resp.Body.Close()
		if v, err := io.ReadAll(resp.Body); err != nil {
			panic(err)
		} else {
			return string(v), err
		}
	} else {
		// Get new token (extended time)
		newAuth := resp.Header.Get("Authorization")
		if len(newAuth) > 0 {
			c.authorization = newAuth
		}

		// read content
		defer resp.Body.Close()
		var payload string
		if body, err := io.ReadAll(resp.Body); err != nil {
			return "", err
		} else {
			payload = string(body)
		}

		if resp.StatusCode >= 300 {
			return payload, fmt.Errorf("invalid call: %s", resp.Status)
		} else {
			return payload, nil
		}
	}
}

// GetUsername access server at that endpoint and returns this content
func (c *ClientSession) GetUsername() (string, error) {
	return c.callEndpoint("GET", CONNECTION_BASE+"self/user/whoami", "")
}

// SetUserPassword changes user password
func (c *ClientSession) SetUserPassword(password string) error {
	_, err := c.callEndpoint("POST", CONNECTION_BASE+"self/user/password", password)
	return err
}

// AddUser is an admin task that adds an user with that username and password.
// NOTE THAT this user has no role at all once created
func (c *ClientSession) AddUser(username, password string) error {
	payload := map[string]string{"name": username, "password": password}
	if body, err := json.Marshal(payload); err != nil {
		return err
	} else if _, err := c.callEndpoint("POST", CONNECTION_BASE+"manage/user/create", string(body)); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes user by login
func (c *ClientSession) DeleteUser(username string) error {
	if len(username) == 0 {
		return errors.New("cannot create user with empty username")
	}

	path := fmt.Sprintf(CONNECTION_BASE+"manage/user/%s/delete", username)
	_, err := c.callEndpoint("DELETE", path, "")
	return err
}

// GetUserRoles gets the resources and linked roles for current user (needs admin or root)
func (c *ClientSession) GetUserRoles(username string) (map[string][]string, error) {
	result := make(map[string][]string)
	if len(username) == 0 {
		return result, errors.New("cannot get roles of empty user")
	}

	path := fmt.Sprintf(CONNECTION_BASE+"manage/user/%s/access/list", username)
	var loaded map[string]any
	if payload, err := c.callEndpoint("GET", path, ""); err != nil {
		return nil, err
	} else if err := json.Unmarshal([]byte(payload), &loaded); err != nil {
		return nil, err
	} else {
		for group, v := range loaded {
			if values, ok := v.([]any); !ok {
				return nil, errors.New("malformed input")
			} else {
				var roles []string
				for _, value := range values {
					if role, ok := value.(string); !ok {
						return nil, errors.New("malformed input")
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

// SetUserRolesForGroups changes user access for a given user to set those roles for those groups
func (c *ClientSession) SetUserRolesForGroups(username string, access map[string][]string) error {
	path := fmt.Sprintf(CONNECTION_BASE+"manage/user/%s/access/edit", username)
	if len(access) == 0 {
		return errors.New("nil input not accepted")
	} else if _, found := access[""]; found {
		return errors.New("empty group not allowed")
	} else if body, err := json.Marshal(access); err != nil {
		return err
	} else if len(username) == 0 {
		return errors.New("empty username not accepted")
	} else if _, err := c.callEndpoint("PUT", path, string(body)); err != nil {
		return err
	}

	return nil
}
