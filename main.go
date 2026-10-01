package main

import (
	"app/handlers"
	"app/storage"
	"context"
	"errors"
	"fmt"
	"net/http"
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

	chatHandler := handlers.ChatHandler{Database: db}
	http.HandleFunc("/api/v1/chats", chatHandler.GetListUserChats)

	defer storage.Close(context.Background(), db)
}
