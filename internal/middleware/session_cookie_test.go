package middleware

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"
)

func TestSessionCookieSecurityAndLifecycle(t *testing.T) {
	cases := []struct {
		host, forwarded string
		tls, secure     bool
	}{
		{host: "localhost:3000"}, {host: "127.0.0.1:8080"}, {host: "[::1]:8080"},
		{host: "chat.example.org:443", secure: true}, {host: "localhost:3000", forwarded: "https", secure: true}, {host: "localhost", tls: true, secure: true},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("GET", "http://"+tc.host+"/", nil)
		req.Host = tc.host
		if tc.forwarded != "" {
			req.Header.Set("X-Forwarded-Proto", tc.forwarded)
		}
		if tc.tls {
			req.TLS = &tls.ConnectionState{}
		}
		if got := secureRequest(req); got != tc.secure {
			t.Errorf("secureRequest(%q) = %t, want %t", tc.host, got, tc.secure)
		}
	}
	req := httptest.NewRequest("GET", "http://localhost/", nil)
	response := httptest.NewRecorder()
	SetSessionCookie(response, req, "session-value")
	cookie := response.Result().Cookies()[0]
	if cookie.Name != SessionCookieName || cookie.Value != "session-value" || !cookie.HttpOnly || cookie.Secure || cookie.MaxAge <= 0 {
		t.Fatalf("session cookie = %#v", cookie)
	}
	response = httptest.NewRecorder()
	ClearSessionCookie(response, req)
	cookie = response.Result().Cookies()[0]
	if cookie.Value != "" || cookie.MaxAge != -1 {
		t.Fatalf("cleared cookie = %#v", cookie)
	}
	if id, err := NewSessionID(); err != nil || id.String() == "" {
		t.Fatalf("NewSessionID() = %s, %v", id, err)
	}
}
