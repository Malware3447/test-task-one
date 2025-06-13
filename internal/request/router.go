package request

import (
	"context"
	"github.com/go-chi/chi/v5"
	"test-task-one/internal/service/crut"
)

type Router struct {
	router  *chi.Mux
	crtServ *crut.Service
}

func NewRouter(crt *crut.Service) *Router {
	return &Router{
		router:  nil,
		crtServ: crt,
	}
}

func (r *Router) Init(ctx context.Context) {
	const op = "router.Init"
	ctx = context.WithValue(ctx, "router", op)

	r.router = chi.NewRouter()

	r.router.Route("/task/v1", func(r chi.Router) {
		r.Route("/good", func(r chi.Router) {
			r.Post("/create/{procjectId}", r.crtServ.CrbeateGood)
		})
	})

}
