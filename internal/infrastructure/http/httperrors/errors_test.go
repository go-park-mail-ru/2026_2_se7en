package httperrors

import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestHTTPErrorHelpers(t *testing.T) {
	cause := errors.New("failure")
	details := []ErrorDetail{{Field: "name", Reason: "required"}}
	items := []*HTTPError{
		NewBadRequest("bad", details), NewUnauthorized("login"), NewForbidden("denied"), NewNotFound("missing"),
		NewConflict("duplicate", details), NewConflictWithCode("CUSTOM", "duplicate", details), NewInternalError("broken", cause),
	}
	for _, item := range items {
		if item.Error() != item.Message || item.ErrorCode() != item.Code || item.HTTPStatus() != item.Status {
			t.Fatalf("error helpers mismatch: %#v", item)
		}
		if item.ToResponse().Message != item.Message {
			t.Fatalf("response mismatch: %#v", item)
		}
	}
	if !errors.Is(NewInternalError("broken", cause), cause) {
		t.Fatal("unwrap did not return cause")
	}
	response := httptest.NewRecorder()
	WriteError(response, items[0])
	if response.Code != items[0].Status || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response: %d %q", response.Code, response.Header().Get("Content-Type"))
	}
}
