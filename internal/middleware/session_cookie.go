package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	SessionCookieName = "session_id"
	SessionTTL        = 30 * 24 * time.Hour
)

func NewSessionID() (uuid.UUID, error) {
	return uuid.NewRandom()
}

func newSessionCookie(value string, maxAge int, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}

func secureRequest(r *http.Request) bool {
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	host := r.Host
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	return host != "localhost" && !net.ParseIP(host).IsLoopback()
}

func SetSessionCookie(w http.ResponseWriter, r *http.Request, id string) {
	http.SetCookie(w, newSessionCookie(id, int(SessionTTL.Seconds()), secureRequest(r)))
}

func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, newSessionCookie("", -1, secureRequest(r)))
}
