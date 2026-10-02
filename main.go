package main

import (
	"app/storage"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const databaseURLEnv = "DATABASE_URL"

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки .env: %v\n", err)
		os.Exit(1)
	}

	db, err := storage.Connect(context.Background(), os.Getenv(databaseURLEnv))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer storage.Close(db)
}
