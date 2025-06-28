package dto

import "time"

type AuditEntryLog struct {
	EventDate        time.Time `json:"date"`
	EventInitiator   string    `json:"initiator"`
	EventType        string    `json:"type"`
	EventDescription string    `json:"description"`
	EventParameters  []string  `json:"parameters"`
}
