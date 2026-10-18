package register

import (
	"net/http"
)

type RegisterDelivery interface {
	HandleRegister(w http.ResponseWriter, r *http.Request)
}
