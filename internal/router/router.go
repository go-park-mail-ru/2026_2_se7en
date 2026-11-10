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
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/csrf"
	usecaseauth "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/auth"
	usecasechat "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/chat"
	usecasecontact "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/contact"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
	usecaseregister "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/register"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func New(db *storage.DB, origins []string, baseLog *slog.Logger, csrfTokens *csrf.HashToken) (http.Handler, *ws.Hub) {
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

	wsHandler := deliverychat.NewWSHandler(hub, chatUsecase, origins, baseLog)

	authDelivery := deliveryauth.NewAuthDelivery(authUsecase, hub, baseLog)
	registerDelivery := deliveryregister.NewRegisterDelivery(registerUsecase, baseLog)
	chatDelivery := deliverychat.NewChatDelivery(chatUsecase, hub, baseLog)
	profileDelivery := deliveryprofile.NewProfileDelivery(profileUsecase, baseLog)
	contactDelivery := deliverycontact.NewContactDelivery(contactUsecase, hub, baseLog)

	api := http.NewServeMux()
	csrfProtection := middleware.NewCSRFProtection(origins, csrfTokens, baseLog)
	auth := func(next http.Handler) http.Handler {
		return middleware.NewAuthMiddleware(sessionRepo, csrfProtection.Handler(next), baseLog)
	}

	api.HandleFunc("GET /health", healthCheck)
	api.Handle("GET /csrf", auth(http.HandlerFunc(csrfProtection.Token)))

	api.Handle("POST /auth/register", tools.NewJSONDecoder(registerDelivery.HandleRegister, baseLog))
	api.Handle("POST /auth/login", tools.NewJSONDecoder(authDelivery.HandleLogin, baseLog))
	api.Handle("GET /auth/me", auth(http.HandlerFunc(authDelivery.HandleMe)))
	api.Handle("POST /auth/logout", auth(http.HandlerFunc(authDelivery.HandleLogout)))

	api.Handle("GET /chats", auth(http.HandlerFunc(chatDelivery.HandleGetListUserChats)))
	api.Handle("GET /chats/{id}", auth(http.HandlerFunc(chatDelivery.HandleGetChatDetail)))
	api.Handle("POST /chats/dialogs", auth(tools.NewJSONDecoder(chatDelivery.HandleCreateDialog, baseLog)))
	api.Handle("POST /chats/{id}/hide", auth(http.HandlerFunc(chatDelivery.HandleHideDialog)))
	api.Handle("POST /chats/groups", auth(tools.NewJSONDecoder(chatDelivery.HandleCreateGroup, baseLog)))
	api.Handle("PATCH /chats/{id}", auth(tools.NewJSONDecoder(chatDelivery.HandleUpdateGroup, baseLog)))
	api.Handle("PUT /chats/{id}/avatar", auth(http.HandlerFunc(chatDelivery.HandleUploadGroupAvatar))) // Из sync-dev
	api.Handle("DELETE /chats/{id}", auth(http.HandlerFunc(chatDelivery.HandleDeleteChat)))
	api.Handle("GET /chats/{id}/members", auth(http.HandlerFunc(chatDelivery.HandleGetListMembers)))
	api.Handle("POST /chats/{id}/members", auth(tools.NewJSONDecoder(chatDelivery.HandleAddMember, baseLog)))
	api.Handle("PATCH /chats/{id}/members/{user_id}/role", auth(tools.NewJSONDecoder(chatDelivery.HandleUpdateMemberRole, baseLog))) // Из sync-dev
	api.Handle("DELETE /chats/{id}/members/{user_id}", auth(http.HandlerFunc(chatDelivery.HandleRemoveMember)))
	api.Handle("DELETE /chats/{id}/leave", auth(http.HandlerFunc(chatDelivery.HandleLeaveGroup)))
	api.Handle("GET /chats/{id}/messages", auth(http.HandlerFunc(chatDelivery.HandleGetListMessages)))
	api.Handle("POST /chats/{id}/messages", auth(tools.NewJSONDecoder(chatDelivery.HandleCreateMessage, baseLog)))
	api.Handle("GET /ws", auth(http.HandlerFunc(wsHandler.HandleWS)))

	api.Handle("GET /contacts", auth(http.HandlerFunc(contactDelivery.HandleGetContactsList)))
	api.Handle("POST /contacts", auth(tools.NewJSONDecoder(contactDelivery.HandleCreateContact, baseLog)))
	api.Handle("DELETE /contacts", auth(tools.NewJSONDecoder(contactDelivery.HandleDeleteContact, baseLog)))

	api.Handle("GET /users/{id}", auth(http.HandlerFunc(profileDelivery.HandleGetPublicProfile))) // Из dev
	api.Handle("GET /profile", auth(http.HandlerFunc(profileDelivery.HandleGetProfile)))
	api.Handle("PATCH /profile", auth(tools.NewJSONDecoder(profileDelivery.HandleUpdateProfile, baseLog)))
	api.Handle("PUT /profile/avatar", auth(http.HandlerFunc(profileDelivery.HandleUploadAvatar)))
	api.Handle("PATCH /profile/password", auth(tools.NewJSONDecoder(profileDelivery.HandleUpdatePassword, baseLog)))
	api.Handle("GET /swagger/*any", httpSwagger.WrapHandler)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", csrfProtection.CheckOrigin(http.StripPrefix("/api/v1", api)))
	mainRouter.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	handler := middleware.RequestIDMiddleware(
		middleware.AccessLogMiddleware(mainRouter, baseLog),
	)

	return handler, hub
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	tools.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
