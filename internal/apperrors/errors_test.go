package apperrors

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestServiceErrorWrapAndUnwrap(t *testing.T) {
	base := errors.New("cause")
	wrapped := Wrap(base)
	if wrapped == nil || !errors.Is(wrapped, base) {
		t.Fatalf("wrapped error does not preserve causes: %v", wrapped)
	}
	var serviceErr *ServiceError
	if wrapped.Error() == "" || !errors.As(wrapped, &serviceErr) || serviceErr.Unwrap() == nil {
		t.Fatalf("invalid wrapped error: %+v", wrapped)
	}
}

func TestWrapDatabase(t *testing.T) {
	if WrapDatabase(nil) != nil {
		t.Fatal("WrapDatabase(nil) should be nil")
	}
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "unique violation", err: &pgconn.PgError{Code: "23505"}},
		{name: "foreign key", err: &pgconn.PgError{Code: "23503"}},
		{name: "unknown database error", err: errors.New("offline")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := WrapDatabase(tc.err)
			var serviceErr *ServiceError
			if !errors.As(got, &serviceErr) || serviceErr.Code != "DATABASE_ERROR" || !errors.Is(got, tc.err) {
				t.Fatalf("WrapDatabase(%v) = %v", tc.err, got)
			}
		})
	}
}
