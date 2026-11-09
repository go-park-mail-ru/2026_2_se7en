package router

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"

	deliveryauth "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/auth"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat"
	deliverychat "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat"
	ws "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat/ws"
	deliverycontact "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/contact"
	deliveryregister "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/register"

	usecaseauth "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/auth"
	usecasechat "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/chat"
	usecasecontact "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/contact"
	usecaseregister "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/register"

	deliveryprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/profile"
	repochat "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/chat"
	repocontact "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/contact"
	repoprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	reposession "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/session"
	repouser "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/user"

	repofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/file"
	repotx "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func New(db *storage.DB, origins []string) (http.Handler, *ws.Hub) {
	hub := ws.NewHub()

	userRepo := repouser.NewUserRepository(db.Pool())
	profileRepo := repoprofile.NewProfileRepository(db.Pool())
	sessionRepo := reposession.NewSessionRepository(db.Pool())
	chatRepo := repochat.NewChatRepository(db.Pool())
	fileRepo := repofile.NewFileRepository(db.Pool())
	txManager := repotx.NewManager(db.Pool())
	contactRepo := repocontact.NewContactRepository(db.Pool())

	authUsecase := usecaseauth.NewAuthUsecase(userRepo, profileRepo, sessionRepo)
	registerUsecase := usecaseregister.NewRegisterUsecase(userRepo, profileRepo)
	chatUsecase := usecasechat.NewChatUsecase(chatRepo)
	profileUsecase := usecaseprofile.NewProfileUsecase(txManager, userRepo, profileRepo, fileRepo)
	contactUsecase := usecasecontact.NewContactUsecase(contactRepo, userRepo)

	wsHandler := chat.NewWSHandler(hub, chatUsecase, origins)

	authDelivery := deliveryauth.NewAuthDelivery(authUsecase, hub)
	registerDelivery := deliveryregister.NewRegisterDelivery(registerUsecase)
	chatDelivery := deliverychat.NewChatDelivery(chatUsecase, hub)
	profileDelivery := deliveryprofile.NewProfileDelivery(profileUsecase)
	contactDelivery := deliverycontact.NewContactDelivery(contactUsecase, hub)

	api := http.NewServeMux()

	api.HandleFunc("GET /health", healthCheck)

	api.Handle("POST /auth/register", tools.NewJSONDecoder(registerDelivery.HandleRegister))
	api.Handle("POST /auth/login", tools.NewJSONDecoder(authDelivery.HandleLogin))

	api.Handle("GET /auth/me", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleMe)))
	api.Handle("POST /auth/logout", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleLogout)))
	api.Handle("GET /chats", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetListUserChats)))
	api.Handle("GET /chats/{id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetChatDetail)))
	api.Handle("POST /chats/dialogs", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleCreateDialog)))
	api.Handle("POST /chats/{id}/hide", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleHideDialog)))
	api.Handle("POST /chats/groups", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleCreateGroup)))
	api.Handle("PATCH /chats/{id}", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleUpdateGroup)))
	api.Handle("PUT /chats/{id}/avatar", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleUploadGroupAvatar)))
	api.Handle("DELETE /chats/{id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleDeleteChat)))
	api.Handle("GET /chats/{id}/members", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetListMembers)))
	api.Handle("POST /chats/{id}/members", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleAddMember)))
	api.Handle("DELETE /chats/{id}/members/{user_id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleRemoveMember)))
	api.Handle("DELETE /chats/{id}/leave", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleLeaveGroup)))
	api.Handle("GET /chats/{id}/messages", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetListMessages)))
	api.Handle("POST /chats/{id}/messages", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleCreateMessage)))

	api.Handle("GET /ws", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(wsHandler.HandleWS)))

	api.Handle("GET /contacts", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(contactDelivery.HandleGetContactsList)))
	api.Handle("POST /contacts", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(contactDelivery.HandleCreateContact)))
	api.Handle("DELETE /contacts", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(contactDelivery.HandleDeleteContact)))

	api.Handle("GET /profile", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleGetProfile)))
	api.Handle("PATCH /profile", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(profileDelivery.HandleUpdateProfile)))
	api.Handle("PUT /profile/avatar", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleUploadAvatar)))
	api.Handle("PATCH /profile/password", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(profileDelivery.HandleUpdatePassword)))

	api.Handle("GET /swagger/*any", httpSwagger.WrapHandler)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
	mainRouter.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	return mainRouter, hub
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	tools.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
