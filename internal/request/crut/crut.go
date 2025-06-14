package crut

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	modelsEvent "test-task-one/internal/models/ch"
	"test-task-one/internal/models/requests"
	natsClient "test-task-one/internal/nats"
	"test-task-one/internal/service/db/ch"
	"test-task-one/internal/service/db/pg"
)

type Crut struct {
	repoPg *pg.Service
	repoCh *ch.Service
	nats   *natsClient.NATSClient
}

type Params struct {
	RepoPg *pg.Service
	RepoCh *ch.Service
	NATS   *natsClient.NATSClient
}

func NewCrut(params Params) CrutHudnler {
	return &Crut{
		repoPg: params.RepoPg,
		repoCh: params.RepoCh,
		nats:   params.NATS,
	}
}

func (c *Crut) ProcessNATSMessages(ctx context.Context) {
	go func() {
		err := c.nats.ProcessMessages("goods.events", func(event *modelsEvent.Event) error {
			err := c.repoCh.LogEvent(ctx, event)
			if err != nil {
				log.Printf("Failed to log event in ClickHouse: %v", err)
				return err
			}
			log.Printf("Event successfully logged in ClickHouse: %v", event.ID)
			return nil
		})
		if err != nil {
			log.Printf("Failed to process NATS messages: %v", err)
		}
	}()
}
func (c *Crut) CreateGood(w http.ResponseWriter, r *http.Request) {
	const op = "router.CreateGood"
	ctx := context.WithValue(r.Context(), "router", op)
	log.Println("Создание новой записи...")

	projectId := chi.URLParam(r, "projectId")
	IdGood, err := strconv.Atoi(projectId)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse project ID: %v", op, err)
		return
	}

	body := requests.Create{}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := c.repoPg.CreateGood(ctx, int32(IdGood), body.Name)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s: Failed to create good: %v", op, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Запись успешно добавлена")
}
