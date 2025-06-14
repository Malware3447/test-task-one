package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"test-task-one/internal/models/ch"
	"test-task-one/internal/models/pg"
	"test-task-one/internal/nats"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryPg struct {
	db   *pgxpool.Pool
	nats *nats.NATSClient
}

func NewRepositoryPg(db *pgxpool.Pool, nats *nats.NATSClient) Repository {
	return &RepositoryPg{
		db:   db,
		nats: nats,
	}
}

func (r *RepositoryPg) CreateGood(ctx context.Context, projectID int32, name string) (*pg.Good, error) {
	// Проверяем, существует ли проект
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}

	// Если проект не существует, создаем новый
	if !exists {
		const insertProject = `INSERT INTO projects (name) VALUES ($1) RETURNING id`
		err := r.db.QueryRow(ctx, insertProject, fmt.Sprintf("Project for good: %s", name)).Scan(&projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert project: %w", err)
		}
	}

	// Теперь вставляем товар
	const insertGood = `INSERT INTO goods (project_id, name, priority)
    VALUES ($1, $2, (SELECT COALESCE(MAX(priority), 0) + 1 FROM goods WHERE project_id = $1))
    RETURNING id, priority, created_at`

	newGood := &pg.Good{
		ProjectID: projectID,
		Name:      name,
	}

	err = r.db.QueryRow(ctx, insertGood, projectID, name).Scan(&newGood.ID, &newGood.Priority, &newGood.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert good: %w", err)
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

func (r *RepositoryPg) GetGood(ctx context.Context, id int32) (*pg.Good, error) {
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
			return nil, fmt.Errorf("good not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get good: %w", err)
	}
	return &good, nil
}

func (r *RepositoryPg) UpdateGood(ctx context.Context, id int32, name, description *string) (*pg.Good, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	good, err := r.GetGood(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		good.Name = *name
	}
	if description != nil {
		good.Description = *description
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
		Description: good.Description,
		Priority:    good.Priority,
		Removed:     good.Removed,
		EventTime:   time.Now(),
	}
	if err := r.nats.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish event: %v\n", err)
	}

	return good, nil
}

func (r *RepositoryPg) MarkAsRemoved(ctx context.Context, id int32) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	good, err := r.GetGood(ctx, id)
	if err != nil {
		return err
	}

	good.Removed = true

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET removed = true 
        WHERE id = $1
    `, id)
	if err != nil {
		return fmt.Errorf("failed to mark as removed: %w", err)
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

	return tx.Commit(ctx)
}

func (r *RepositoryPg) ReprioritizeGood(ctx context.Context, id int32, newPriority int32) ([]pg.Good, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	good, err := r.GetGood(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET priority = $1 
        WHERE id = $2
    `, newPriority, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update priority: %w", err)
	}

	_, err = tx.Exec(ctx, `
        UPDATE goods 
        SET priority = priority + 1 
        WHERE project_id = $1 
          AND priority >= $2 
          AND id != $3
    `, good.ProjectID, newPriority, id)
	if err != nil {
		return nil, fmt.Errorf("failed to shift priorities: %w", err)
	}

	rows, err := tx.Query(ctx, `
        SELECT id, priority 
        FROM goods 
        WHERE project_id = $1 
          AND priority >= $2
        ORDER BY priority
    `, good.ProjectID, newPriority)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated priorities: %w", err)
	}
	defer rows.Close()

	var updated []pg.Good
	for rows.Next() {
		var g pg.Good
		err := rows.Scan(&g.ID, &g.Priority)
		if err != nil {
			return nil, fmt.Errorf("failed to scan priority: %w", err)
		}
		updated = append(updated, g)
	}

	return updated, tx.Commit(ctx)
}

func (r *RepositoryPg) ListGoods(ctx context.Context, limit, offset int32) ([]pg.Good, int, int, error) {
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM goods").Scan(&total)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	var removedCount int
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM goods WHERE removed = true").Scan(&removedCount)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get removed count: %w", err)
	}

	rows, err := r.db.Query(ctx, `
        SELECT id, project_id, name, description, priority, removed, created_at 
        FROM goods 
        ORDER BY priority 
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get goods list: %w", err)
	}
	defer rows.Close()

	var goods []pg.Good
	for rows.Next() {
		var g pg.Good
		err := rows.Scan(
			&g.ID,
			&g.ProjectID,
			&g.Name,
			&g.Description,
			&g.Priority,
			&g.Removed,
			&g.CreatedAt,
		)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to scan good: %w", err)
		}
		goods = append(goods, g)
	}

	return goods, total, removedCount, nil
}
