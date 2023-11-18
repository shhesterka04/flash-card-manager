package kafka

import (
	"time"
)

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Query     string    `json:"raw_query"`
}

func NewEvent(eventType, query string) Event {
	return Event{
		Timestamp: time.Now(),
		Type:      eventType,
		Query:     query,
	}
}
