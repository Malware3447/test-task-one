package crut

import (
	"net/http"
	"test-task-one/internal/request/crut"
)

type Service struct {
	crt crut.CrutHudnler
}

func NewService(crt crut.CrutHudnler) *Service {
	return &Service{crt: crt}
}

func (s *Service) CreateGood(w http.ResponseWriter, r *http.Request) {
	s.crt.CreateGood(w, r)
}
