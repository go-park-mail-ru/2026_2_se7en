package middleware

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/csrf"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
)

func TestCSRFProtection(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokens, err := csrf.NewHMACHashToken("test-secret-with-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	origin := "https://app.example"
	protection := NewCSRFProtection([]string{origin}, tokens, log)
	userID, sessionID := uuid.New(), uuid.New()
	ctx := NewAuthContext(NewSessionContext(context.Background(), sessionID), userID)
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		called = false
		protection.CheckOrigin(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, "/", nil))
		if !called {
			t.Fatalf("safe method %s was blocked", method)
		}
	}
	blocked := httptest.NewRecorder()
	badOriginReq := httptest.NewRequest(http.MethodPost, "/", nil)
	badOriginReq.Header.Set("Origin", "https://evil.example")
	protection.CheckOrigin(next).ServeHTTP(blocked, badOriginReq)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("untrusted origin status = %d", blocked.Code)
	}
	allowed := httptest.NewRequest(http.MethodPost, "/", nil)
	allowed.Header.Set("Origin", origin)
	called = false
	protection.CheckOrigin(next).ServeHTTP(httptest.NewRecorder(), allowed)
	if !called {
		t.Fatal("trusted origin was blocked")
	}

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		called = false
		protection.Handler(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, "/", nil))
		if !called {
			t.Fatalf("safe request %s was blocked by CSRF token check", method)
		}
	}
	noSession := httptest.NewRecorder()
	protection.Handler(next).ServeHTTP(noSession, httptest.NewRequest(http.MethodPost, "/", nil))
	if noSession.Code != http.StatusUnauthorized {
		t.Fatalf("missing session status = %d", noSession.Code)
	}

	requestWithContext := func() *http.Request { return httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx) }
	missingToken := httptest.NewRecorder()
	protection.Handler(next).ServeHTTP(missingToken, requestWithContext())
	if missingToken.Code != http.StatusForbidden {
		t.Fatalf("missing token status = %d", missingToken.Code)
	}
	duplicate := httptest.NewRecorder()
	duplicateRequest := requestWithContext()
	duplicateRequest.Header.Add(csrfHeaderName, "one")
	duplicateRequest.Header.Add(csrfHeaderName, "two")
	protection.Handler(next).ServeHTTP(duplicate, duplicateRequest)
	if duplicate.Code != http.StatusForbidden {
		t.Fatalf("duplicate token status = %d", duplicate.Code)
	}
	invalid := httptest.NewRecorder()
	invalidRequest := requestWithContext()
	invalidRequest.Header.Set(csrfHeaderName, "invalid")
	protection.Handler(next).ServeHTTP(invalid, invalidRequest)
	if invalid.Code != http.StatusForbidden {
		t.Fatalf("invalid token status = %d", invalid.Code)
	}

	token := tokens.Create(&models.Session{ID: sessionID, UserID: userID}, time.Now().Add(time.Hour).Unix())
	valid := httptest.NewRecorder()
	validRequest := requestWithContext()
	validRequest.Header.Set(csrfHeaderName, token)
	called = false
	protection.Handler(next).ServeHTTP(valid, validRequest)
	if !called {
		t.Fatal("valid token was rejected")
	}

	tokenResponse := httptest.NewRecorder()
	tokenRequest := httptest.NewRequest(http.MethodGet, "/csrf", nil).WithContext(ctx)
	tokenRequest.Header.Set("Origin", origin)
	protection.Token(tokenResponse, tokenRequest)
	if tokenResponse.Code != http.StatusOK || tokenResponse.Header().Get("Cache-Control") != "no-store" || !strings.Contains(tokenResponse.Body.String(), "csrf_token") {
		t.Fatalf("token response = %d %s", tokenResponse.Code, tokenResponse.Body.String())
	}
	unauthorizedToken := httptest.NewRecorder()
	protection.Token(unauthorizedToken, httptest.NewRequest(http.MethodGet, "/csrf", nil))
	if unauthorizedToken.Code != http.StatusUnauthorized {
		t.Fatalf("token without session = %d", unauthorizedToken.Code)
	}
	blockedToken := httptest.NewRecorder()
	blockedReq := httptest.NewRequest(http.MethodGet, "/csrf", nil).WithContext(ctx)
	blockedReq.Header.Set("Origin", "https://evil.example")
	protection.Token(blockedToken, blockedReq)
	if blockedToken.Code != http.StatusForbidden {
		t.Fatalf("untrusted token origin = %d", blockedToken.Code)
	}
}

type authRepoStub struct {
	user              *models.SessionUser
	getErr, extendErr error
}

func (s authRepoStub) GetSessionUser(context.Context, uuid.UUID) (*models.SessionUser, error) {
	return s.user, s.getErr
}
func (s authRepoStub) ExtendSession(context.Context, uuid.UUID, time.Time) error { return s.extendErr }

func TestAuthMiddlewareFailureBranches(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("next should not run") })
	validSession := uuid.New()
	call := func(repo authRepoStub, cookie string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: cookie})
		}
		response := httptest.NewRecorder()
		NewAuthMiddleware(repo, next, log).ServeHTTP(response, req)
		return response
	}
	if got := call(authRepoStub{}, ""); got.Code != http.StatusUnauthorized {
		t.Fatalf("missing cookie = %d", got.Code)
	}
	if got := call(authRepoStub{}, "bad-id"); got.Code != http.StatusUnauthorized {
		t.Fatalf("bad cookie = %d", got.Code)
	}
	if got := call(authRepoStub{getErr: apperrors.ErrSessionNotFound}, validSession.String()); got.Code != http.StatusUnauthorized {
		t.Fatalf("expired session = %d", got.Code)
	}
	if got := call(authRepoStub{getErr: errors.New("db")}, validSession.String()); got.Code != http.StatusInternalServerError {
		t.Fatalf("lookup failure = %d", got.Code)
	}
	if got := call(authRepoStub{}, validSession.String()); got.Code != http.StatusInternalServerError {
		t.Fatalf("nil session user = %d", got.Code)
	}
	if got := call(authRepoStub{user: &models.SessionUser{UserID: uuid.New()}, extendErr: apperrors.ErrSessionNotFound}, validSession.String()); got.Code != http.StatusUnauthorized {
		t.Fatalf("expired on extend = %d", got.Code)
	}
	if got := call(authRepoStub{user: &models.SessionUser{UserID: uuid.New()}, extendErr: errors.New("db")}, validSession.String()); got.Code != http.StatusInternalServerError {
		t.Fatalf("extend failure = %d", got.Code)
	}
}
