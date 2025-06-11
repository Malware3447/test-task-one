package pg

import (
	"context"
	"test-task-one/internal/models"
)

type Repository interface {
	CreateGood(ctx context.Context, projectID int32, name string) (*models.Good, error)
	GetGood(ctx context.Context, id int32) (*models.Good, error)
	UpdateGood(ctx context.Context, id int32, name, description *string) (*models.Good, error)
	MarkAsRemoved(ctx context.Context, id int32) error
	ReprioritizeGood(ctx context.Context, id int32, newPriority int32) ([]models.Good, error)
	ListGoods(ctx context.Context, limit, offset int32) ([]models.Good, int, int, error)
}
