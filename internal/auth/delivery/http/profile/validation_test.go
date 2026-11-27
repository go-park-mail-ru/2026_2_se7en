package profile

import "testing"

func TestUpdateProfileRequestValidationAndNormalization(t *testing.T) {
	email, phone, first, last := "  TEST@example.com ", "  +79991234567  ", " Alice ", " Smith "
	req := &UpdateProfileRequest{Email: &email, PhoneNumber: &phone, FirstName: &first, LastName: &last}
	if details := req.IsValid(); len(details) != 0 {
		t.Fatalf("details = %+v", details)
	}
	if *req.Email != "test@example.com" || *req.PhoneNumber != "+79991234567" {
		t.Fatalf("request not normalized: %+v", req)
	}
}

func TestUpdateProfileRequestReturnsFieldErrors(t *testing.T) {
	email, nickname, first, phone := "bad", "!", "", "letters"
	req := &UpdateProfileRequest{Email: &email, Nickname: &nickname, FirstName: &first, PhoneNumber: &phone}
	if details := req.IsValid(); len(details) < 4 {
		t.Fatalf("details = %+v, want validation errors", details)
	}
}

func TestUpdatePasswordRequestValidation(t *testing.T) {
	if got := (&UpdatePasswordRequest{OldPassword: "old", NewPassword: "goodpass123"}).IsValid(); len(got) != 0 {
		t.Fatalf("details = %+v", got)
	}
	if got := (&UpdatePasswordRequest{}).IsValid(); len(got) != 2 {
		t.Fatalf("details = %+v, want old and new password errors", got)
	}
}
