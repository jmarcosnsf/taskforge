package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"taskforge/db/sqlc"
	"taskforge/internal/handler"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to execute code", "error", err)
		return
	}
}

func run() error {
	godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil{
		return err
	}

	rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

	repository := sqlc.New((pool))

	handler := handler.NewHandler(repository, os.Getenv("JWT_SECRET"), rdb)
	
	s := http.Server{
		ReadTimeout: 10 * time.Second,
		IdleTimeout: time.Minute,
		WriteTimeout: 10 * time.Second,
		Addr: ":8080",
		Handler: handler,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}