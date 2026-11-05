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

	repochat "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/chat"
	repoprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	reposession "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/session"
	repouser "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/user"
)

func New(db *storage.DB) http.Handler {
	userRepo := repouser.NewUserRepository(db.Pool())
	profileRepo := repoprofile.NewProfileRepository(db.Pool())
	sessionRepo := reposession.NewSessionRepository(db.Pool())
	chatRepo := repochat.NewChatRepository(db.Pool())

	authUsecase := usecaseauth.NewAuthUsecase(userRepo, profileRepo, sessionRepo)
	registerUsecase := usecaseregister.NewRegisterUsecase(userRepo, profileRepo)
	chatUsecase := usecasechat.NewChatUsecase(chatRepo)

	authDelivery := deliveryauth.NewAuthDelivery(authUsecase)
	registerDelivery := deliveryregister.NewRegisterDelivery(registerUsecase)
	chatDelivery := deliverychat.NewChatDelivery(chatUsecase)

	api := http.NewServeMux()

	api.Handle("POST /auth/register", tools.NewJSONDecoder(registerDelivery.HandleRegister))
	api.Handle("POST /auth/login", tools.NewJSONDecoder(authDelivery.HandleLogin))

	api.Handle("GET /auth/me", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleMe)))
	api.Handle("POST /auth/logout", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(authDelivery.HandleLogout)))
	api.Handle("GET /chats", middleware.NewAuthMiddleware(sessionRepo, http.HandlerFunc(chatDelivery.HandleListUserChats)))

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return mainRouter
}
