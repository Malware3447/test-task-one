package pg

import (
	"context"
	"test-task-one/internal/db/pg"
	models "test-task-one/internal/models/pg"
	"test-task-one/internal/models/responses"
)

type Service struct {
	repo pg.Repository
}

func NewService(repo pg.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateGood(ctx context.Context, projectID int32, name string) (*models.Good, error) {
	return s.repo.CreateGood(ctx, projectID, name)
}

func (s *Service) GetGood(ctx context.Context, id int32) (*models.Good, bool, error) {
	return s.repo.GetGood(ctx, id)
}

func (s *Service) GetProject(ctx context.Context, projectID int32) (bool, error) {
	return s.repo.GetProject(ctx, projectID)
}

func (s *Service) UpdateGood(ctx context.Context, id int32, name, description *string) (*models.Good, error) {
	return s.repo.UpdateGood(ctx, id, name, description)
}

func (s *Service) MarkAsRemoved(ctx context.Context, id int32) (responses.Remove, error) {
	return s.repo.MarkAsRemoved(ctx, id)
}

func (s *Service) ReprioritizeGood(ctx context.Context, id int32, newPriority int32) (responses.Reprioritize, error) {
	return s.repo.ReprioritizeGood(ctx, id, newPriority)
}

func (s *Service) ListGoods(ctx context.Context, limit, offset int32) (responses.List, error) {
	return s.repo.ListGoods(ctx, limit, offset)
}
