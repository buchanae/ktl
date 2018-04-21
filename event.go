package ktl

import (
	"time"
)

const (
	Start EventType = iota
	Stop
)

// EventType enumerates the types of events available.
type EventType int

//go:generate enumer -type=EventType -text

// Event describes an event occurring during the lifetime of a step,
// usually due to a state change, such as "start" or "stop".
// Events are used while processing step drivers.
type Event struct {
	Type      EventType `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	// Processed is set to true when the event has been successfully
	// processed by the step driver.
	Processed bool `json:"processed"`
}

// NewEvent creates a new event of the given type, with the timestamp
// set to now.
func NewEvent(t EventType) *Event {
	return &Event{Type: t, CreatedAt: time.Now()}
}
