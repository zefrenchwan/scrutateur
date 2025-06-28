package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

const CONNECTION_BASE = "http://localhost:3000/"

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

// GetCurrentUserGroups lists current groups and roles for user
func (c *ClientSession) GetCurrentUserGroups() (map[string][]string, error) {
	var body any
	if resp, err := c.callEndpoint("GET", CONNECTION_BASE+"self/groups/list", ""); err != nil {
		return nil, err
	} else if regexp.MustCompile(`\A\s*\z`).MatchString(resp) {
		return nil, nil
	} else if err := json.Unmarshal([]byte(resp), &body); err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	// parse result as a map[string][]string
	if m, ok := body.(map[string]any); !ok {
		return nil, errors.New("invalid input, not a map")
	} else if len(m) == 0 {
		return nil, nil
	} else {
		for name, v := range m {
			var groupRoles []string
			if values, ok := v.([]any); !ok {
				return nil, errors.New("invalid input,value is not a slice")
			} else if len(values) == 0 {
				return nil, errors.New("invalid input, no role for group")
			} else {
				for _, val := range values {
					if role, ok := val.(string); !ok {
						return nil, errors.New("invalid input, role is not a string")
					} else {
						groupRoles = append(groupRoles, role)
					}
				}

				result[name] = groupRoles
			}
		}
	}

	return result, nil
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

// SetUserRolesForFeatures changes user access for a given user to set those roles for those features
func (c *ClientSession) SetUserRolesForFeatures(username string, access map[string][]string) error {
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

// CreateGroupOfUsers creates a group and add current user in that group
func (c *ClientSession) CreateGroupOfUsers(groupName string) error {
	message, err := c.callEndpoint("POST", CONNECTION_BASE+"groups/create/"+groupName, "")
	if err != nil {
		fmt.Println("ERROR: " + message)
	}

	return err
}

// UpsertUserInGroup changes user roles in a group (or add user in the group with those roles)
func (c *ClientSession) UpsertUserInGroup(userName, groupName string, roles []string) error {
	if body, err := json.Marshal(roles); err != nil {
		return err
	} else if message, err := c.callEndpoint("PUT", CONNECTION_BASE+"groups/"+groupName+"/upsert/user/"+userName, string(body)); err != nil {
		fmt.Println("ERROR: " + message)
		return err
	}

	return nil
}

// RevokeUserInGroup revokes an user in a group
func (c *ClientSession) RevokeUserInGroup(userName, groupName string) error {
	if message, err := c.callEndpoint("DELETE", CONNECTION_BASE+"groups/"+groupName+"/revoke/user/"+userName, ""); err != nil {
		fmt.Println("ERROR: " + message)
		return err
	}

	return nil
}

// DeleteGroupOfUsers deletes a group by name
func (c *ClientSession) DeleteGroupOfUsers(groupName string) error {
	message, err := c.callEndpoint("DELETE", CONNECTION_BASE+"groups/delete/"+groupName, "")
	if err != nil {
		fmt.Println("ERROR: " + message)
	}

	return err
}

// Audit displays audit logs between two moments
func (c *ClientSession) Audit(from, to time.Time) (string, error) {
	url := CONNECTION_BASE + "audits/display"
	var empty time.Time
	var counter int
	if empty.Compare(from) < 0 {
		url = url + "?from=" + from.Format("20060102")
		counter = 1
	}

	if to.Compare(empty) > 0 {
		if counter != 0 {
			url = url + "&to="
		} else {
			url = url + "?to="
		}

		url = url + to.Format("20060102")
		counter = counter + 1
	}

	// for two dates, test if they make sense
	if counter >= 2 {
		if from.Compare(to) > 0 {
			return "", errors.New("period of time is invalid")
		}
	}

	return c.callEndpoint("GET", url, "")
}
