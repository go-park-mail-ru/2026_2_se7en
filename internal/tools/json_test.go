package tools

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
)

type jsonTestRequest struct {
	Name string `json:"name"`
}

func (r *jsonTestRequest) IsValid() []httperrors.ErrorDetail {
	if r.Name == "" {
		return []httperrors.ErrorDetail{{Field: "name", Reason: "required"}}
	}
	return nil
}

func TestJSONDecoderResponses(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewJSONDecoder(func(_ http.ResponseWriter, _ *http.Request, req *jsonTestRequest) *httperrors.HTTPError {
		if req.Name == "handler-error" {
			return httperrors.NewConflict("duplicate", nil)
		}
		return nil
	}, log)
	cases := []struct {
		name, body string
		want       int
	}{
		{"success", `{"name":"ok"}`, 200}, {"invalid JSON", `{`, 400}, {"validation", `{}`, 400}, {"handler error", `{"name":"handler-error"}`, 409},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tc.want {
				t.Fatalf("status = %d, want %d", response.Code, tc.want)
			}
		})
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(""))
	req.Body = nil
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != 400 {
		t.Fatalf("nil body status = %d", response.Code)
	}
}
