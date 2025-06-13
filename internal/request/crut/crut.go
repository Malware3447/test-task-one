package crut

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"test-task-one/internal/models/requests"
	"test-task-one/internal/nats"
	"test-task-one/internal/service/db/ch"
	"test-task-one/internal/service/db/pg"
)

type Crut struct {
	repoPg *pg.Service
	repoCh *ch.Service
	nats   *nats.NATSClient
}

type Params struct {
	RepoPg *pg.Service
	RepoCh *ch.Service
	NATS   *nats.NATSClient
}

func NewCrut(params Params) CrutHudnler {
	return &Crut{
		repoPg: params.RepoPg,
		repoCh: params.RepoCh,
		nats:   params.NATS,
	}
}

func (c *Crut) CreateGood(w http.ResponseWriter, r *http.Request) {
	const op = "router.CreateGood"
	ctx := context.WithValue(r.Context(), "router", op)
	log.Println("Создание новой записи...")

	projectId := chi.URLParam(r, "projectId")
	IdGood, err := strconv.Atoi(projectId)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	body := requests.Create{}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := c.repoPg.CreateGood(ctx, int32(IdGood), body.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Запись успешно добавлена")
}
