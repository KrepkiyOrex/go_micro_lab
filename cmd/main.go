package main

import (
	"database/sql"
	"log"
	"net"
	"time"

	"go_micro_lab/internal/repository/postgres"

	_ "github.com/lib/pq"
	grpclib "google.golang.org/grpc"
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

	postgresRepo := postgres.New(db)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpclib.NewServer()

	log.Println("🚀 gRPC server running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
