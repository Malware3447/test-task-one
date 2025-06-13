package ch

import "time"

type Event struct {
	ID          int32     `json:"id"`
	ProjectID   int32     `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Priority    int32     `json:"priority"`
	Removed     bool      `json:"removed"`
	EventTime   time.Time `json:"event_time"`
}
