package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go_micro_lab/internal/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `INSERT INTO tasks (id, payload, status, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5)`
	
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Payload, task.Status, task.CreatedAt, task.UpdatedAt)
	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `SELECT id, payload, status, created_at, updated_at FROM tasks WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	
	var task domain.Task
	err := row.Scan(&task.ID, &task.Payload, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("task not found")
		}
		return nil, err
	}
	
	return &task, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3`
	
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("task not found")
	}
	
	return nil
}