package auth

import "net/http"

type AuthDelivery interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleMe(w http.ResponseWriter, r *http.Request)
	HandleLogout(w http.ResponseWriter, r *http.Request)
}
