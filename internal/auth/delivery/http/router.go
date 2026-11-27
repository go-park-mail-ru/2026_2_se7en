package http

import (
	"log/slog"
	"net/http"

	authdelivery "github.com/go-park-mail-ru/2026_2_se7en/internal/auth/delivery/http/auth"
	contactdelivery "github.com/go-park-mail-ru/2026_2_se7en/internal/auth/delivery/http/contact"
	profiledelivery "github.com/go-park-mail-ru/2026_2_se7en/internal/auth/delivery/http/profile"
	registerdelivery "github.com/go-park-mail-ru/2026_2_se7en/internal/auth/delivery/http/register"
	usersdelivery "github.com/go-park-mail-ru/2026_2_se7en/internal/auth/delivery/http/users"
	httpjson "github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/json"
)

type AuthMiddleware func(http.Handler) http.Handler

type RouterDependencies struct {
	AuthMiddleware AuthMiddleware
	CSRFToken      http.HandlerFunc
	Logger         *slog.Logger

	Auth     *authdelivery.AuthDeliveryImpl
	Register *registerdelivery.RegisterDeliveryImpl
	Contact  *contactdelivery.ContactDeliveryImpl
	Profile  *profiledelivery.ProfileDeliveryImpl
	Users    *usersdelivery.UsersDeliveryImpl
}

func RegisterRoutes(mux *http.ServeMux, deps RouterDependencies) {
	mux.Handle("GET /csrf", deps.AuthMiddleware(http.HandlerFunc(deps.CSRFToken)))

	mux.Handle("POST /auth/register", httpjson.NewJSONDecoder(deps.Register.HandleRegister, deps.Logger))
	mux.Handle("POST /auth/login", httpjson.NewJSONDecoder(deps.Auth.HandleLogin, deps.Logger))
	mux.Handle("GET /auth/me", deps.AuthMiddleware(http.HandlerFunc(deps.Auth.HandleMe)))
	mux.Handle("POST /auth/logout", deps.AuthMiddleware(http.HandlerFunc(deps.Auth.HandleLogout)))

	mux.Handle("GET /contacts", deps.AuthMiddleware(http.HandlerFunc(deps.Contact.HandleGetContactsList)))
	mux.Handle("POST /contacts", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Contact.HandleCreateContact, deps.Logger)))
	mux.Handle("DELETE /contacts", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Contact.HandleDeleteContact, deps.Logger)))

	mux.Handle("GET /users", deps.AuthMiddleware(http.HandlerFunc(deps.Users.HandleGetUsersList)))
	mux.Handle("GET /users/{id}", deps.AuthMiddleware(http.HandlerFunc(deps.Profile.HandleGetPublicProfile)))
	mux.Handle("GET /profile", deps.AuthMiddleware(http.HandlerFunc(deps.Profile.HandleGetProfile)))
	mux.Handle("PATCH /profile", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Profile.HandleUpdateProfile, deps.Logger)))
	mux.Handle("PUT /profile/avatar", deps.AuthMiddleware(http.HandlerFunc(deps.Profile.HandleUploadAvatar)))
	mux.Handle("PATCH /profile/password", deps.AuthMiddleware(httpjson.NewJSONDecoder(deps.Profile.HandleUpdatePassword, deps.Logger)))
}
