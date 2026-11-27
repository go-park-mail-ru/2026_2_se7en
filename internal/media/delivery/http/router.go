package http

import http "net/http"

func RegisterRoutes(mux *http.ServeMux, uploadDir string) {
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
}
