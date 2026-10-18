package router

import (
	"net/http"

	"app/middleware"
	"app/storage"
	"app/utils"

	deliveryauth "app/delivery/auth"
	deliverychat "app/delivery/chat"
	deliveryregister "app/delivery/register"

	usecaseauth "app/usecase/auth"
	usecasechat "app/usecase/chat"
	usecaseregister "app/usecase/register"

	repochat "app/repository/chat"
	repoprofile "app/repository/profile"
	reposession "app/repository/session"
	repouser "app/repository/user"
)

func New(db *storage.DB) http.Handler {
	userRepo := repouser.NewUserRepository(db.Pool())
	profileRepo := repoprofile.NewProfileRepository(db.Pool())
	sessionRepo := reposession.NewSessionRepository(db.Pool())
	chatRepo := repochat.NewChatRepository(db.Pool())

	authUsecase := usecaseauth.NewAuthUsecase(userRepo, profileRepo, sessionRepo)
	registerUsecase := usecaseregister.NewRegisterUsecase(userRepo, profileRepo, sessionRepo)
	chatUsecase := usecasechat.NewChatUsecase(chatRepo)

	authDelivery := deliveryauth.NewAuthDelivery(authUsecase)
	registerDelivery := deliveryregister.NewRegisterDelivery(registerUsecase)
	chatDelivery := deliverychat.NewChatDelivery(chatUsecase)

	api := http.NewServeMux()

	api.HandleFunc("GET /health", healthCheck)

	api.Handle("POST /auth/register", http.HandlerFunc(registerDelivery.HandleRegister))
	api.Handle("POST /auth/login", http.HandlerFunc(authDelivery.HandleLogin))

	api.Handle("GET /auth/me", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleMe)))
	api.Handle("POST /auth/logout", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleLogout)))
	api.Handle("GET /chats", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleListUserChats)))

	api.HandleFunc("GET /swagger", swaggerUI)
	api.Handle("GET /specs/", http.StripPrefix("/specs/", http.FileServer(http.Dir("specs"))))

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return mainRouter
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
