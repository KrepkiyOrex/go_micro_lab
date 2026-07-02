package ports

import (
	"context"
	"go_micro_lab/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}