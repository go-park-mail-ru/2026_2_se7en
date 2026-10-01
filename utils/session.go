package utils

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	SessionCookieName = "session_id"
	SessionTTL = 30 * 24 * time.Hour
)

func NewSessionID() (uuid.UUID, error) {
	return uuid.NewRandom()
}

func SetSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName,
		Value: id,
		Path: "/",
		HttpOnly: true,
		Secure: false, // DEUBG
		SameSite: http.SameSiteLaxMode,
		MaxAge: int(SessionTTL.Seconds()),
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName,
		Value: "",
		Path: "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge: -1,
	})
}
