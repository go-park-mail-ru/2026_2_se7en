package router

import (
	"encoding/json"
	"net/http"
)

func New() http.Handler {
	api := http.NewServeMux()

	api.HandleFunc("GET /health", healthCheck)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("api/v1/", http.StripPrefix("api/v1/", api))

	return mainRouter
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
