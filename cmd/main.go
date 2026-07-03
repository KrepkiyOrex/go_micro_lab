package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	"go_micro_lab/internal/api/grpc"
	"go_micro_lab/internal/repository/postgres"
	redisrepo "go_micro_lab/internal/repository/redis"
	"go_micro_lab/internal/usecase"

	_ "github.com/lib/pq"
	redislib "github.com/redis/go-redis/v9"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/go_micro_lab?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(1 * time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	rdb := redislib.NewClient(&redislib.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}
	log.Println("✅ Connected to Redis")

	postgresRepo := postgres.New(db)
	cachRepo := redisrepo.New(rdb)

	taskUsecase := usecase.NewTaskUsecase(postgresRepo, cachRepo, nil)

	// создаем gRPC сервер и привязываем к нему нашу бизнес логику
	// "адаптер" превращает gRPC запросы в вызовы usecase
	taskServer := grpc.NewTaskServer(taskUsecase)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpclib.NewServer()
	grpc.RegisterTaskServiceServer(grpcServer, taskServer)
	reflection.Register(grpcServer)

	log.Println("🚀 gRPC server running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
