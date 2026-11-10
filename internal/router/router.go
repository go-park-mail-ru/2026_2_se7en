package router

import (
	"log/slog"
	"net/http"

	deliveryauth "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/auth"
	deliverychat "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat"
	ws "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/chat/ws"
	deliverycontact "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/contact"
	deliveryprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/profile"
	deliveryregister "github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/register"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	repochat "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/chat"
	repocontact "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/contact"
	repofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/file"
	repoprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	reposession "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/session"
	repotx "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	repouser "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/user"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	usecaseauth "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/auth"
	usecasechat "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/chat"
	usecasecontact "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/contact"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
	usecaseregister "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/register"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func New(db *storage.DB, origins []string, baseLog *slog.Logger) (http.Handler, *ws.Hub) {
	hub := ws.NewHub(baseLog)

	userRepo := repouser.NewUserRepository(db.Pool(), baseLog)
	profileRepo := repoprofile.NewProfileRepository(db.Pool(), baseLog)
	sessionRepo := reposession.NewSessionRepository(db.Pool(), baseLog)
	chatRepo := repochat.NewChatRepository(db.Pool(), baseLog)
	fileRepo := repofile.NewFileRepository(db.Pool(), baseLog)
	txManager := repotx.NewManager(db.Pool(), baseLog)
	contactRepo := repocontact.NewContactRepository(db.Pool(), baseLog)

	authUsecase := usecaseauth.NewAuthUsecase(userRepo, profileRepo, sessionRepo, baseLog)
	registerUsecase := usecaseregister.NewRegisterUsecase(userRepo, profileRepo, baseLog)
	chatUsecase := usecasechat.NewChatUsecase(chatRepo, baseLog)
	profileUsecase := usecaseprofile.NewProfileUsecase(txManager, userRepo, profileRepo, fileRepo, baseLog)
	contactUsecase := usecasecontact.NewContactUsecase(contactRepo, userRepo, baseLog)

	wsHandler := chat.NewWSHandler(hub, chatUsecase, origins, baseLog)

	authDelivery := deliveryauth.NewAuthDelivery(authUsecase, hub, baseLog)
	registerDelivery := deliveryregister.NewRegisterDelivery(registerUsecase, baseLog)
	chatDelivery := deliverychat.NewChatDelivery(chatUsecase, hub, baseLog)
	profileDelivery := deliveryprofile.NewProfileDelivery(profileUsecase, baseLog)
	contactDelivery := deliverycontact.NewContactDelivery(contactUsecase, baseLog)

	api := http.NewServeMux()

	api.HandleFunc("GET /health", healthCheck)

	api.HandleFunc("GET /health", healthCheck)

	api.Handle("POST /auth/register", tools.NewJSONDecoder(registerDelivery.HandleRegister, baseLog))
	api.Handle("POST /auth/login", tools.NewJSONDecoder(authDelivery.HandleLogin, baseLog))
	api.Handle("GET /auth/me", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleMe), baseLog))
	api.Handle("POST /auth/logout", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleLogout), baseLog))

	api.Handle("GET /chats", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetListUserChats), baseLog))
	api.Handle("GET /chats/{id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetChatDetail), baseLog))
	api.Handle("POST /chats/dialogs", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleCreateDialog, baseLog), baseLog))
	api.Handle("POST /chats/{id}/hide", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleHideDialog), baseLog))
	api.Handle("POST /chats/groups", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleCreateGroup, baseLog), baseLog))
	api.Handle("PATCH /chats/{id}", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleUpdateGroup, baseLog), baseLog))
	api.Handle("PUT /chats/{id}/avatar", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleUploadGroupAvatar), baseLog)) // Из sync-dev
	api.Handle("DELETE /chats/{id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleDeleteChat), baseLog))
	api.Handle("GET /chats/{id}/members", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetListMembers), baseLog))
	api.Handle("POST /chats/{id}/members", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleAddMember, baseLog), baseLog))
	api.Handle("PATCH /chats/{id}/members/{user_id}/role", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleUpdateMemberRole, baseLog), baseLog)) // Из sync-dev
	api.Handle("DELETE /chats/{id}/members/{user_id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleRemoveMember), baseLog))
	api.Handle("DELETE /chats/{id}/leave", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleLeaveGroup), baseLog))
	api.Handle("GET /chats/{id}/messages", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleGetListMessages), baseLog))
	api.Handle("POST /chats/{id}/messages", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(chatDelivery.HandleCreateMessage, baseLog), baseLog))

	api.Handle("GET /ws", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(wsHandler.HandleWS), baseLog))

	api.Handle("GET /contacts", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(contactDelivery.HandleGetContactsList), baseLog))
	api.Handle("POST /contacts", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(contactDelivery.HandleCreateContact, baseLog), baseLog))
	api.Handle("DELETE /contacts", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(contactDelivery.HandleDeleteContact, baseLog), baseLog))

	api.Handle("GET /users/{id}", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleGetPublicProfile), baseLog)) // Из dev
	api.Handle("GET /profile", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleGetProfile), baseLog))
	api.Handle("PATCH /profile", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(profileDelivery.HandleUpdateProfile, baseLog), baseLog))
	api.Handle("PUT /profile/avatar", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(profileDelivery.HandleUploadAvatar), baseLog))
	api.Handle("PATCH /profile/password", middleware.NewAuthMiddleware(sessionRepo, tools.NewJSONDecoder(profileDelivery.HandleUpdatePassword, baseLog), baseLog))

	api.Handle("GET /swagger/*any", httpSwagger.WrapHandler)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
	mainRouter.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	handler := middleware.RequestIDMiddleware(
		middleware.AccessLogMiddleware(mainRouter, baseLog),
	)

	return handler, hub
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	tools.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
