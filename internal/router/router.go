package router

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"

	deliveryauth "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/auth"
	deliverychat "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat"
	deliveryregister "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/register"

	usecaseauth "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/auth"
	usecasechat "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/chat"
	usecaseregister "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/register"

	deliveryprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/profile"
	repochat "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/chat"
	repoprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	reposession "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/session"
	repouser "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/user"

	repofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/file"
	repotx "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
)

func New(db *storage.DB) http.Handler {
	userRepo := repouser.NewUserRepository(db.Pool())
	profileRepo := repoprofile.NewProfileRepository(db.Pool())
	sessionRepo := reposession.NewSessionRepository(db.Pool())
	chatRepo := repochat.NewChatRepository(db.Pool())
	fileRepo := repofile.NewFileRepository(db.Pool())
	txManager := repotx.NewManager(db.Pool())

	authUsecase := usecaseauth.NewAuthUsecase(userRepo, profileRepo, sessionRepo)
	registerUsecase := usecaseregister.NewRegisterUsecase(userRepo, profileRepo, sessionRepo)
	chatUsecase := usecasechat.NewChatUsecase(chatRepo)
	profileUsecase := usecaseprofile.NewProfileUsecase(txManager, userRepo, profileRepo, fileRepo)

	authDelivery := deliveryauth.NewAuthDelivery(authUsecase)
	registerDelivery := deliveryregister.NewRegisterDelivery(registerUsecase)
	chatDelivery := deliverychat.NewChatDelivery(chatUsecase)
	profileDelivery := deliveryprofile.NewProfileDelivery(profileUsecase)

	api := http.NewServeMux()

	api.HandleFunc("GET /health", healthCheck)

	api.Handle("POST /auth/register", tools.NewJSONDecoder(registerDelivery.HandleRegister))
	api.Handle("POST /auth/login", tools.NewJSONDecoder(authDelivery.HandleLogin))

	api.Handle("GET /auth/me", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleMe)))
	api.Handle("POST /auth/logout", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleLogout)))
	api.Handle("GET /chats", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleListUserChats)))

	api.Handle("GET /profile", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleGetProfile)))
	api.Handle("PATCH /profile", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(profileDelivery.HandleUpdateProfile)))
	api.Handle("PUT /profile/avatar", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleUploadAvatar)))
	api.Handle("PATCH /profile/password", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(profileDelivery.HandleUpdatePassword)))

	api.HandleFunc("GET /swagger", swaggerUI)
	api.HandleFunc("GET /swagger/", swaggerUI)
	api.HandleFunc("GET /swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
	mainRouter.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	return mainRouter
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	tools.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
