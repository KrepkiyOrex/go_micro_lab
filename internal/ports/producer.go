package ports

import "context"

type EventProducer interface {
	PublishTaskCreated(ctx context.Context, taskID string) error
	Close() error
}