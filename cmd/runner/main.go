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
	"test-task-one/internal/app"
	"test-task-one/internal/config"
	"test-task-one/internal/db/ch"
	"test-task-one/internal/db/pg"
	"test-task-one/internal/nats"
	"test-task-one/internal/request"
	"test-task-one/internal/request/crut"
	serviceCrut "test-task-one/internal/service/crut"
	serviceCh "test-task-one/internal/service/db/ch"
	servicePg "test-task-one/internal/service/db/pg"
)

func main() {
	const op = "cmd.runner.main"
	cfg, _ := configo.MustLoad[config.Config]()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "main", op)

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

	natsClient, err := nats.NewNATSClient("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	log.Println("NATS успешно запущен")

	log.Println("Сервер успешно запущен")

	repoPg := pg.NewRepositoryPg(poolPg, natsClient)
	repoCh := ch.NewRepositoryCh(poolCh)

	pgService := servicePg.NewService(repoPg)
	chService := serviceCh.NewService(repoCh)

	crutParams := crut.Params{
		RepoPg: pgService,
		RepoCh: chService,
		NATS:   natsClient,
	}

	crutHandler := crut.NewCrut(crutParams)

	crutService := serviceCrut.NewService(crutHandler)

	router := request.NewRouter(crutService)

	App := app.NewApp(router)

	App.Init(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-quit:
		log.Println("Завершение работы сервиса")
	}

	log.Println("Сервис успешно завершил работу")
}
