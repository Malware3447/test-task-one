package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"log"
	"test-task-one/internal/models/ch"
	"test-task-one/internal/models/pg"
	"test-task-one/internal/models/responses"
	"test-task-one/internal/nats"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryPg struct {
	db    *pgxpool.Pool
	nats  *nats.NATSClient
	redis *redis.Client
}

type Params struct {
	Db    *pgxpool.Pool
	Nats  *nats.NATSClient
	Redis *redis.Client
}

func NewRepositoryPg(params *Params) Repository {
	return &RepositoryPg{
		db:    params.Db,
		nats:  params.Nats,
		redis: params.Redis,
	}
}

func (r *RepositoryPg) CreateGood(ctx context.Context, projectID int32, name string) (*pg.Good, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1) FOR UPDATE", projectID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}

	if !exists {
		const insertProject = `INSERT INTO projects (name) VALUES ($1) RETURNING id`
		err := tx.QueryRow(ctx, insertProject, fmt.Sprintf("Project for good: %s", name)).Scan(&projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert project: %w", err)
		}
	}

	const insertGood = `INSERT INTO goods (project_id, name, priority)
    VALUES ($1, $2, (SELECT COALESCE(MAX(priority), 0) + 1 FROM goods WHERE project_id = $1 FOR UPDATE))
    RETURNING id, priority, created_at`

	newGood := &pg.Good{
		ProjectID: projectID,
		Name:      name,
	}

	err = tx.QueryRow(ctx, insertGood, projectID, name).Scan(&newGood.ID, &newGood.Priority, &newGood.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert good: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	event := &ch.Event{
		ID:        newGood.ID,
		ProjectID: newGood.ProjectID,
		Name:      newGood.Name,
		Priority:  newGood.Priority,
		EventTime: newGood.CreatedAt,
	}
	if err := r.nats.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish event: %v\n", err)
	} else {
		log.Println("Событие успешно отправлено в NATS")
	}

	return newGood, nil
}

func (r *RepositoryPg) GetProject(ctx context.Context, projectID int32) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return exists, nil
}

func (r *RepositoryPg) GetGood(ctx context.Context, id int32) (*pg.Good, bool, error) {
	var good pg.Good
	err := r.db.QueryRow(ctx, `
        SELECT id, project_id, name, description, priority, removed, created_at 
        FROM goods 
        WHERE id = $1 
        FOR UPDATE
    `, id).Scan(
		&good.ID,
		&good.ProjectID,
		&good.Name,
		&good.Description,
		&good.Priority,
		&good.Removed,
		&good.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, fmt.Errorf("good not found: %w", err)
		}
		return nil, false, fmt.Errorf("failed to get good: %w", err)
	}
	return &good, true, nil
}

func (r *RepositoryPg) UpdateGood(ctx context.Context, id int32, name, description *string) (*pg.Good, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	good, _, err := r.GetGood(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		good.Name = *name
	}
	if description != nil {
		good.Description = description
	}

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET name = $1, description = $2 
        WHERE id = $3
    `, good.Name, good.Description, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update good: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	event := &ch.Event{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: *good.Description,
		Priority:    good.Priority,
		Removed:     good.Removed,
		EventTime:   time.Now(),
	}
	if err := r.nats.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish event: %v\n", err)
	}

	cacheKey := fmt.Sprintf("good:%d", id)
	jsonData, err := json.Marshal(good)
	if err == nil {
		if err := r.redis.Set(ctx, cacheKey, jsonData, 0).Err(); err != nil {
			fmt.Printf("Failed to cache updated good: %v\n", err)
		}
	}

	return good, nil
}

func (r *RepositoryPg) MarkAsRemoved(ctx context.Context, id int32) (responses.Remove, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return responses.Remove{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	good, _, err := r.GetGood(ctx, id)
	if err != nil {
		return responses.Remove{}, err
	}

	good.Removed = true

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET removed = true 
        WHERE id = $1
    `, id)
	if err != nil {
		return responses.Remove{}, fmt.Errorf("failed to mark as removed: %w", err)
	}

	resp := responses.Remove{
		Id:         good.ID,
		CampaignId: good.ProjectID,
		Removed:    true,
	}

	event := &ch.Event{
		ID:        good.ID,
		ProjectID: good.ProjectID,
		Name:      good.Name,
		Removed:   true,
		EventTime: time.Now(),
	}
	if err := r.nats.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish event: %v\n", err)
	}

	return resp, tx.Commit(ctx)
}

func (r *RepositoryPg) ReprioritizeGood(ctx context.Context, id int32, newPriority int32) (responses.Reprioritize, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return responses.Reprioritize{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET priority = $1 
        WHERE id = $2
    `, newPriority, id)
	if err != nil {
		return responses.Reprioritize{}, fmt.Errorf("failed to update priority for id %d: %w", id, err)
	}

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET priority = priority + 1 
        WHERE id < $1
    `, id)
	if err != nil {
		return responses.Reprioritize{}, fmt.Errorf("failed to update priorities for other goods: %w", err)
	}

	rows, err := tx.Query(ctx, `
        SELECT id, priority 
        FROM goods 
        WHERE id <= $1 
        ORDER BY id
    `, id)
	if err != nil {
		return responses.Reprioritize{}, fmt.Errorf("failed to fetch updated priorities: %w", err)
	}
	defer rows.Close()

	var updatedPriorities []responses.Priorities
	for rows.Next() {
		var p responses.Priorities
		if err := rows.Scan(&p.Id, &p.Priority); err != nil {
			return responses.Reprioritize{}, fmt.Errorf("failed to scan priority: %w", err)
		}
		updatedPriorities = append(updatedPriorities, p)

		event := &ch.Event{
			ID:        p.Id,
			Priority:  p.Priority,
			EventTime: time.Now(),
		}
		if err := r.nats.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Failed to publish event for id %d: %v\n", p.Id, err)
		} else {
			log.Printf("Событие успешно отправлено в NATS для id %d\n", p.Id)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return responses.Reprioritize{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := responses.Reprioritize{Priorities: updatedPriorities}

	return result, nil
}
func (r *RepositoryPg) ListGoods(ctx context.Context, limit, offset int32) (responses.List, error) {
	cacheKey := fmt.Sprintf("goods:list:%d:%d", limit, offset)

	cachedData, err := r.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var cachedList responses.List
		if err := json.Unmarshal(cachedData, &cachedList); err == nil {
			return cachedList, nil
		}
	}

	var total int32
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM goods").Scan(&total)
	if err != nil {
		return responses.List{}, fmt.Errorf("failed to get total count: %w", err)
	}

	var removedCount int32
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM goods WHERE removed = true").Scan(&removedCount)
	if err != nil {
		return responses.List{}, fmt.Errorf("failed to get removed count: %w", err)
	}

	rows, err := r.db.Query(ctx, `
        SELECT id, project_id, name, description, priority, removed, created_at 
        FROM goods 
        ORDER BY priority 
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return responses.List{}, fmt.Errorf("failed to get goods list: %w", err)
	}
	defer rows.Close()

	var goods []responses.Goods
	for rows.Next() {
		var g responses.Goods
		err := rows.Scan(
			&g.Id,
			&g.ProjectId,
			&g.Name,
			&g.Description,
			&g.Priority,
			&g.Romoved,
			&g.CreatedAt,
		)
		if err != nil {
			return responses.List{}, fmt.Errorf("failed to scan good: %w", err)
		}
		goods = append(goods, g)
	}

	resp := responses.List{
		Meta: responses.Meta{
			Total:   total,
			Removed: removedCount,
			Limit:   limit,
			Offset:  offset,
		},
		Goods: goods,
	}

	jsonData, err := json.Marshal(resp)
	if err == nil {
		r.redis.Set(ctx, cacheKey, jsonData, time.Minute)
	}

	return resp, nil
}
