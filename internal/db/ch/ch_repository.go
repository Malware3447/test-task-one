package ch

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"test-task-one/internal/models/ch"
)

type RepositoryCh struct {
	db driver.Conn
}

func NewRepositoryCh(db driver.Conn) Repository {
	return &RepositoryCh{db: db}
}

// LogEvent записывает событие в ClickHouse
func (r *RepositoryCh) LogEvent(ctx context.Context, event *ch.Event) error {
	const query = `
        INSERT INTO events (
            id, 
            project_id, 
            name, 
            description, 
            priority, 
            removed, 
            event_time
        ) VALUES (
            @id, 
            @project_id, 
            @name, 
            @description, 
            @priority, 
            @removed, 
            @event_time
        )
    `

	err := r.db.Exec(ctx, query,
		clickhouse.Named("id", event.ID),
		clickhouse.Named("project_id", event.ProjectID),
		clickhouse.Named("name", event.Name),
		clickhouse.Named("description", event.Description),
		clickhouse.Named("priority", event.Priority),
		clickhouse.Named("removed", event.Removed),
		clickhouse.Named("event_time", event.EventTime),
	)
	if err != nil {
		return fmt.Errorf("failed to log event: %w", err)
	}

	return nil
}
