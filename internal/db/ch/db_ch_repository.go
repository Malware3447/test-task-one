package ch

import (
	"context"
	"test-task-one/internal/models/ch"
)

type Repository interface {
	LogEvent(ctx context.Context, event *ch.Event) error
}
