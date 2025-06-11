package main

import (
	"context"
	"fmt"
	"github.com/Malware3447/configo"
	"github.com/Malware3447/sch"
	"github.com/Malware3447/spg"
	"log"
	"os"
	"os/signal"
	"syscall"
	"test-task-one/internal/config"
	"test-task-one/internal/db/ch"
	"test-task-one/internal/db/pg"
)

func main() {
	const op = "cmd.runner.main"
	cfg, _ := configo.MustLoad[config.Config]()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "logger", op)

	poolPg, err := spg.NewClient(ctx, &cfg.DatabasePg)
	if err != nil {
		log.Println(fmt.Errorf("ошибка при запуске Postgres: %s", err))
		panic(err)
	}
	log.Println("Postgres успешно запущен")

	poolCh, err := sch.NewClient(ctx, &cfg.DatabaseCh)
	if err != nil {
		log.Println(fmt.Errorf("ошибка при запуске ClickHouse: %s", err))
		panic(err)
	}
	log.Println("ClickHouse успешно запущен")

	_ = pg.NewRepositoryPg(poolPg)
	_ = ch.NewRepositoryCh(poolCh)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-quit:
		log.Println("Завершение работы сервиса")
	}

	log.Println("Сервис успешно завершил работу")
}
