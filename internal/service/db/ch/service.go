package ch

import (
	"context"
	"test-task-one/internal/db/ch"
	modelsCh "test-task-one/internal/models/ch"
)

type Service struct {
	repo ch.Repository
}

func NewService(repo ch.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LogEvent(ctx context.Context, event *modelsCh.Event) error {
	return s.repo.LogEvent(ctx, event)
}
