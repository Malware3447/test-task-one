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
	projId, err := strconv.Atoi(projectId)
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

	response, err := c.repoPg.CreateGood(ctx, int32(projId), body.Name)
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

func (c *Crut) GoodUpdate(w http.ResponseWriter, r *http.Request) {
	const op = "router.GoodUpdate"
	ctx := context.WithValue(r.Context(), "router", op)
	log.Println("Обновление записи...")

	projectId := chi.URLParam(r, "projectId")
	projId, err := strconv.Atoi(projectId)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse project ID: %v", op, err)
		return
	}

	goodId := chi.URLParam(r, "id")
	Id, err := strconv.Atoi(goodId)
	if err != nil {
		http.Error(w, "Invalid good ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse good ID: %v", op, err)
		return
	}
	exits, _ := c.repoPg.GetProject(ctx, int32(projId))
	if !exits {
		http.Error(w, "Project not found", http.StatusNotFound)
		log.Println("Проект не найден")

		notFound := struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Details map[string]string `json:"details"`
		}{
			Code:    "1",
			Message: "errors.common.notFound",
			Details: map[string]string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notFound)
		return
	}

	_, exits, _ = c.repoPg.GetGood(ctx, int32(Id))
	if !exits {
		http.Error(w, "Good not found", http.StatusNotFound)
		log.Println("Запись не найдена")

		notFound := struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Details map[string]string `json:"details"`
		}{
			Code:    "1",
			Message: "errors.common.notFound",
			Details: map[string]string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notFound)
		return
	}

	body := requests.Update{}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := c.repoPg.UpdateGood(ctx, int32(Id), &body.Name, &body.Description)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s: Failed to update good: %v", op, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Запись успешно обновлена")
}

func (c *Crut) GoodRemove(w http.ResponseWriter, r *http.Request) {
	const op = "router.GoodRemove"
	ctx := context.WithValue(r.Context(), "router", op)
	log.Println("Удаление записи...")

	projectId := chi.URLParam(r, "projectId")
	projId, err := strconv.Atoi(projectId)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse project ID: %v", op, err)
		log.Println("Проект не найден")

		notFound := struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Details map[string]string `json:"details"`
		}{
			Code:    "1",
			Message: "errors.common.notFound",
			Details: map[string]string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notFound)
		return
	}

	goodId := chi.URLParam(r, "id")
	Id, err := strconv.Atoi(goodId)
	if err != nil {
		http.Error(w, "Invalid good ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse good ID: %v", op, err)
		log.Println("Запись не найдена")

		notFound := struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Details map[string]string `json:"details"`
		}{
			Code:    "1",
			Message: "errors.common.notFound",
			Details: map[string]string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notFound)
		return
	}
	exits, _ := c.repoPg.GetProject(ctx, int32(projId))
	if !exits {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	_, exits, _ = c.repoPg.GetGood(ctx, int32(Id))
	if !exits {
		http.Error(w, "Good not found", http.StatusNotFound)
		return
	}

	response, err := c.repoPg.MarkAsRemoved(ctx, int32(Id))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s: Failed to update good: %v", op, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Запись успешно удалена")
}

func (c *Crut) GoodList(w http.ResponseWriter, r *http.Request) {
	const op = "router.GoodList"
	ctx := context.WithValue(r.Context(), "router", op)
	log.Println("Отправляем записи...")

	limitUrl := chi.URLParam(r, "limit")
	limit, err := strconv.Atoi(limitUrl)
	if err != nil {
		http.Error(w, "Invalid good ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse good ID: %v", op, err)
		return
	}

	offsetUrl := chi.URLParam(r, "offset")
	offset, err := strconv.Atoi(offsetUrl)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse project ID: %v", op, err)
		return
	}

	response, err := c.repoPg.ListGoods(ctx, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s: Failed to update good: %v", op, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Все записи отправлены")
}

func (c *Crut) ReprioritizeGood(w http.ResponseWriter, r *http.Request) {
	const op = "router.GoodRemove"
	ctx := context.WithValue(r.Context(), "router", op)
	log.Println("Обновляем приоритеты...")

	projectId := chi.URLParam(r, "projectId")
	projId, err := strconv.Atoi(projectId)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse project ID: %v", op, err)
		return
	}

	goodId := chi.URLParam(r, "id")
	Id, err := strconv.Atoi(goodId)
	if err != nil {
		http.Error(w, "Invalid good ID", http.StatusBadRequest)
		log.Printf("%s: Failed to parse good ID: %v", op, err)
		return
	}
	exits, _ := c.repoPg.GetProject(ctx, int32(projId))
	if !exits {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	_, exits, _ = c.repoPg.GetGood(ctx, int32(Id))
	if !exits {
		http.Error(w, "Good not found", http.StatusNotFound)
		return
	}

	body := requests.Reprioritize{}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := c.repoPg.ReprioritizeGood(ctx, int32(Id), body.NewPriority)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s: Failed to update good: %v", op, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Все приоритеты обновлены")
}
