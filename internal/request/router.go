package request

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
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

	r.router.Route("/task/v1", func(router chi.Router) {
		router.Route("/good", func(router chi.Router) {
			router.Post("/create/{projectId}/", r.crtServ.CreateGood)
		})
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", 8080), r.router); err != nil {
			panic(fmt.Sprintf("%v: %v", op, err))
		}
	}()
}
