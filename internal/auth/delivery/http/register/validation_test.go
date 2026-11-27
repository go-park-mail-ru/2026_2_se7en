package register

import "testing"

func TestRegisterRequestValidationNormalizesOptionalFields(t *testing.T) {
	phone, last := "  ", " Smith "
	req := &RegisterRequest{Email: "user@example.com", Password: "validpassword", PhoneNumber: &phone, Nickname: "user", FirstName: " Alice ", LastName: &last}
	if details := req.IsValid(); len(details) != 0 {
		t.Fatalf("details = %+v", details)
	}
	if req.PhoneNumber != nil || req.FirstName != "Alice" || req.LastName == nil || *req.LastName != "Smith" {
		t.Fatalf("request not normalized: %+v", req)
	}
}

func TestRegisterRequestValidationReportsInvalidInput(t *testing.T) {
	req := &RegisterRequest{Email: "bad", Password: "x", Nickname: "!", FirstName: ""}
	if details := req.IsValid(); len(details) < 4 {
		t.Fatalf("details = %+v, want multiple validation errors", details)
	}
}
