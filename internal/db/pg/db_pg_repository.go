package pg

import (
	"context"
	"test-task-one/internal/models/pg"
)

type Repository interface {
	CreateGood(ctx context.Context, projectID int32, name string) (*pg.Good, error)
	GetGood(ctx context.Context, id int32) (*pg.Good, error)
	UpdateGood(ctx context.Context, id int32, name, description *string) (*pg.Good, error)
	MarkAsRemoved(ctx context.Context, id int32) error
	ReprioritizeGood(ctx context.Context, id int32, newPriority int32) ([]pg.Good, error)
	ListGoods(ctx context.Context, limit, offset int32) ([]pg.Good, int, int, error)
}
