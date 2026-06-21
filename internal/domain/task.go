package domain

import "time"

type Task struct {
	ID        string
	Payload   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// статусы задач
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusError      = "error"
)

func NewTask(id, payload string) *Task {
	now := time.Now()
	return &Task{
		ID:        id,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}