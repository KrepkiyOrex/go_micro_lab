package usecase

import (
	"context"
	"encoding/json"
	"go_micro_lab/internal/domain"
	"go_micro_lab/internal/ports"
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	cacheTTL  = 5 * time.Minute
	dbTimeout = 500 * time.Millisecond
)

type TaskUsecase struct {
	repo     ports.TaskRepository
	cache    ports.CacheRepository
	producer ports.EventProducer
}

func NewTaskUsecase(
	repo ports.TaskRepository,
	cache ports.CacheRepository,
	producer ports.EventProducer,
) *TaskUsecase {
	return &TaskUsecase{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

func (uc *TaskUsecase) CreateTask(ctx context.Context, payload string) (*domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	id := uuid.New().String()
	task := domain.NewTask(id, payload)

	if err := uc.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	go func() {
		taskJSON, _ := json.Marshal(task)
		if err := uc.cache.Set(context.Background(), "task:"+task.ID, string(taskJSON), cacheTTL); err != nil {
			log.Printf("❌ Failed to cache task %s: %v", task.ID, err)
		} else {
			log.Printf("✅ Cached task %s in Redis", task.ID)
		}
	}()

	if uc.producer != nil {
		go func() {
			if err := uc.producer.PublishTaskCreated(context.Background(), task.ID); err != nil {
				log.Printf("❌ Failed to send event for task %s: %v", task.ID, err)
			}
		}()
	}

	return task, nil
}

func (uc *TaskUsecase) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	if uc.cache == nil {
		ctx, cancel := context.WithTimeout(ctx, dbTimeout)
		defer cancel()
		return uc.repo.GetByID(ctx, id)
	}

	cacheKey := "task:" + id

	cached, err := uc.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var task domain.Task
		if err := json.Unmarshal([]byte(cached), &task); err == nil {
			log.Printf("📦 Cache HIT for task %s (from Redis)", id)
			return &task, nil
		}
	}

	log.Printf("💾 Cache MISS for task %s (from PostgreSQL)", id)

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	go func() {
		taskJSON, _ := json.Marshal(task)
		if err := uc.cache.Set(context.Background(), cacheKey, string(taskJSON), cacheTTL); err != nil {
			log.Printf("❌ Failed to cache task %s: %v", task.ID, err)
		} else {
			log.Printf("✅ Cached task %s in Redis (from PostgreSQL)", task.ID)
		}
	}()

	return task, nil
}

func (uc *TaskUsecase) UpdateTaskStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}
