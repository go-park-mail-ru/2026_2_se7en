package chat

import "net/http"

type ChatDelivery interface {
	HandleListUserChats(w http.ResponseWriter, r *http.Request)
}
