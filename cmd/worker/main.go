package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"go_micro_lab/internal/repository/postgres"
	"go_micro_lab/internal/usecase"

	_ "github.com/lib/pq"
	"github.com/IBM/sarama"
)

type TaskCreatedEvent struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/go_micro_lab?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("✅ Worker connected to PostgreSQL")

	repo := postgres.New(db)
	taskUsecase := usecase.NewTaskUsecase(repo, nil, nil)

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("task-events", 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	log.Println("🚀 Worker is listening for task events...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var event TaskCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("❌ Failed to parse message: %v", err)
				continue
			}

			log.Printf("📥 Received event: task_id=%s", event.TaskID)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := taskUsecase.UpdateTaskStatus(ctx, event.TaskID, "done"); err != nil {
				log.Printf("❌ Failed to update task %s: %v", event.TaskID, err)
			} else {
				log.Printf("✅ Task %s updated to done", event.TaskID)
			}
			cancel()

		case err := <-partitionConsumer.Errors():
			log.Printf("❌ Consumer error: %v", err)

		case <-sigChan:
			log.Println("👋 Worker stopped")
			return
		}
	}
}