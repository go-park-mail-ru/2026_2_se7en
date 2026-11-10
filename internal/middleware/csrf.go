package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/csrf"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
)

const (
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenTTL   = 24 * time.Hour
)

type CSRFProtection struct {
	origins []string
	log     *slog.Logger
	tokens  *csrf.HashToken
}

func NewCSRFProtection(origins []string, tokens *csrf.HashToken, baseLog *slog.Logger) *CSRFProtection {
	return &CSRFProtection{
		origins: slices.Clone(origins),
		log:     logger.Scope(baseLog, logger.LayerMiddleware, logger.ModuleCSRF),
		tokens:  tokens,
	}
}

func (p *CSRFProtection) CheckOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if !p.trustedOrigin(r) {
			p.log.WarnContext(r.Context(), "request origin rejected", logger.KeyMethod, r.Method)
			apperrors.WriteError(w, apperrors.NewForbidden("Request origin is not allowed"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (p *CSRFProtection) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		session, ok := csrfSession(r)
		if !ok {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session"))
			return
		}
		headers := r.Header.Values(csrfHeaderName)
		if len(headers) != 1 {
			p.log.WarnContext(r.Context(), "CSRF token missing or duplicated")
			apperrors.WriteError(w, apperrors.NewForbidden("Invalid CSRF token"))
			return
		}
		if err := p.tokens.Check(session, headers[0]); err != nil {
			p.log.WarnContext(r.Context(), "CSRF token rejected", logger.Error(err))
			apperrors.WriteError(w, apperrors.NewForbidden("Invalid CSRF token"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (p *CSRFProtection) Token(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if (len(r.Header.Values("Origin")) > 0 || len(r.Header.Values("Referer")) > 0) && !p.trustedOrigin(r) {
		p.log.WarnContext(r.Context(), "CSRF token origin rejected")
		apperrors.WriteError(w, apperrors.NewForbidden("Request origin is not allowed"))
		return
	}

	session, ok := csrfSession(r)
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session"))
		return
	}
	tools.WriteJSON(w, http.StatusOK, map[string]string{
		"csrf_token": p.tokens.Create(session, time.Now().Add(csrfTokenTTL).Unix()),
	})
}

func csrfSession(r *http.Request) (*models.Session, bool) {
	sessionID, sessionOK := SessionIDFromContext(r.Context())
	userID, userOK := UserIDFromContext(r.Context())
	if !sessionOK || !userOK {
		return nil, false
	}
	return &models.Session{ID: sessionID, UserID: userID}, true
}

func (p *CSRFProtection) trustedOrigin(r *http.Request) bool {
	origin := csrfRequestOrigin(r)
	return origin != "" && origin != "null" && slices.Contains(p.origins, origin)
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
