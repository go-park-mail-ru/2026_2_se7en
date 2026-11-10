package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"slices"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
)

func CSRFMiddleware(next http.Handler, origins []string, baseLog *slog.Logger) http.Handler {
	log := logger.Scope(baseLog, logger.LayerMiddleware, logger.ModuleCSRF)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		origin := csrfRequestOrigin(r)
		if origin == "" || origin == "null" || !slices.Contains(origins, origin) {
			log.WarnContext(r.Context(), "request origin rejected", logger.KeyMethod, r.Method)
			apperrors.WriteError(w, apperrors.NewForbidden("Request origin is not allowed"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func csrfRequestOrigin(r *http.Request) string {
	if origins := r.Header.Values("Origin"); len(origins) > 0 {
		if len(origins) != 1 {
			return ""
		}
		return origins[0]
	}

	referers := r.Header.Values("Referer")
	if len(referers) != 1 {
		return ""
	}
	ref, err := url.Parse(referers[0])
	if err != nil || ref.Host == "" || ref.User != nil || (ref.Scheme != "http" && ref.Scheme != "https") {
		return ""
	}
	return ref.Scheme + "://" + ref.Host
}
