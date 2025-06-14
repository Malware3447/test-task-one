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

func (s *Service) GoodUpdate(w http.ResponseWriter, r *http.Request) {
	s.crt.GoodUpdate(w, r)
}

func (s *Service) GoodRemove(w http.ResponseWriter, r *http.Request) {
	s.crt.GoodRemove(w, r)
}

func (s *Service) GoodList(w http.ResponseWriter, r *http.Request) {
	s.crt.GoodList(w, r)
}

func (s *Service) ReprioritizeGood(w http.ResponseWriter, r *http.Request) {
	s.crt.ReprioritizeGood(w, r)
}
