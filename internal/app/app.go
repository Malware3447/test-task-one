package app

import (
	"context"
	"log"
	"test-task-one/internal/request"
)

type App struct {
	router *request.Router
}

func NewApp(router *request.Router) *App {
	return &App{router: router}
}

func (a *App) Init(ctx context.Context) {
	const op = "app.Init"
	ctx = context.WithValue(ctx, "app", op)

	go a.router.Init(ctx)

	log.Println("Роутер инициализирован")
}
