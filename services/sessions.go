package services

import "encoding/json"

type Session struct {
	CurrentUser string `json:"user"`
}

func NewSessionForUser(user string) Session {
	return Session{CurrentUser: user}
}

func (s *Session) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

func SessionLoad(value []byte) (Session, error) {
	var result Session
	err := json.Unmarshal(value, &result)
	return result, err
}
