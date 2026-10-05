package router

import (
	"app/handlers"
	"app/middleware"
	"app/storage"
	"app/utils"
	"net/http"
)

func New(db *storage.DB) http.Handler {
	api := http.NewServeMux()

	api.HandleFunc("GET /health", healthCheck)

	auth := handlers.NewAuthHandler(db, db)
	registration := handlers.NewHandler(db)
	chats := handlers.NewChatHandler(db)
	api.Handle("POST /auth/register", http.HandlerFunc(registration.RegisterUser))
	api.Handle("POST /auth/login", http.HandlerFunc(auth.Login))
	api.Handle("GET /auth/me", middleware.NewAuthMiddleware(db, http.HandlerFunc(auth.CurrentUser)))
	api.Handle("POST /auth/logout", middleware.NewAuthMiddleware(db, http.HandlerFunc(auth.Logout)))
	api.Handle("GET /chats", middleware.NewAuthMiddleware(db, http.HandlerFunc(chats.GetListUserChats)))

	api.HandleFunc("GET /swagger", swaggerUI)
	api.Handle("GET /specs/", http.StripPrefix("/specs/", http.FileServer(http.Dir("specs"))))

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return mainRouter
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
