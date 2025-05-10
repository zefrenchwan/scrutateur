package dto

import "encoding/json"

// Session contains a serializable set of values for a session
type Session struct {
	CurrentUser string `json:"user"`
}

// NewSessionForUser builds an empty session for a user
func NewSessionForUser(user string) Session {
	return Session{CurrentUser: user}
}

// Serialize to store content in a cache
func (s *Session) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// SessionLoad deserializes a session
func SessionLoad(value []byte) (Session, error) {
	var result Session
	err := json.Unmarshal(value, &result)
	return result, err
}
