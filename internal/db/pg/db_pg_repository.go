package pg

import (
	"context"
	"test-task-one/internal/models/pg"
	"test-task-one/internal/models/responses"
)

type Repository interface {
	CreateGood(ctx context.Context, projectID int32, name string) (*pg.Good, error)
	GetGood(ctx context.Context, id int32) (*pg.Good, bool, error)
	GetProject(ctx context.Context, projectID int32) (bool, error)
	UpdateGood(ctx context.Context, id int32, name, description *string) (*pg.Good, error)
	MarkAsRemoved(ctx context.Context, id int32) (responses.Remove, error)
	ReprioritizeGood(ctx context.Context, id int32, newPriority int32) (responses.Reprioritize, error)
	ListGoods(ctx context.Context, limit, offset int32) (responses.List, error)
}
