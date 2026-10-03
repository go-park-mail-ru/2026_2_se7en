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

func newSessionCookie(value string, maxAge int) *http.Cookie {
    return &http.Cookie{
        Name:     SessionCookieName,
        Value:    value,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        MaxAge:   maxAge,
    }
}

func SetSessionCookie(w http.ResponseWriter, id string) {
    http.SetCookie(w, newSessionCookie(id, int(SessionTTL.Seconds())))
}

func ClearSessionCookie(w http.ResponseWriter) {
    http.SetCookie(w, newSessionCookie("", -1))
}
