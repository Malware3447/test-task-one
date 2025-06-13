package pg

import "time"

type Good struct {
	ID          int32     `json:"id"`
	ProjectID   int32     `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int32     `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"created_at"`
}
