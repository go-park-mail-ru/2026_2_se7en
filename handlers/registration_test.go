package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"app/models"
)

const (
	testEmail        = "user@example.com"
	testPassword     = "password123"
	testPhone        = "+1234567890"
	testPhoneRaw     = " " + testPhone + " "
	testNickname     = "anna_1"
	testFirstName    = "Anna"
	testFirstNameRaw = "  " + testFirstName + "  "
	testLastName     = "Smith"
	testLastNameRaw  = "  " + testLastName + "  "
)

const (
	invalidPhone     = "+12 34"
	shortPassword    = "short"
	longPassword     = "12345678901234567"
	shortNickname    = "ab"
	longNickname     = "aaaaaaaaaaaaaaaaa"
	invalidNickname  = "anna!"
	longEmailPrefix  = 251
	longName         = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	databaseErrorMsg = "database unavailable"
)

type mockRegistrationDB struct {
	emailExists      bool
	nickExists       bool
	emailCheckCalled bool
	checkEmailErr    error
	checkNickErr     error
	createUserErr    error
	createProfileErr error
	userID           uuid.UUID
	createdEmail     string
	createdHash      string
	createdPhone     *string
	createdNickname  string
	createdFirstName string
	createdLastName  *string
	deletedUserIDs   []uuid.UUID
}

func (m *mockRegistrationDB) CheckEmailExist(context.Context, string) (bool, error) {
	m.emailCheckCalled = true
	return m.emailExists, m.checkEmailErr
}

func (m *mockRegistrationDB) CheckNicknameExist(context.Context, string) (bool, error) {
	return m.nickExists, m.checkNickErr
}

func (m *mockRegistrationDB) CreateUser(_ context.Context, email, passwordHash string, phone *string) (*models.User, error) {
	m.createdEmail = email
	m.createdHash = passwordHash
	m.createdPhone = phone
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}
	return &models.User{ID: m.userID, Email: email, PasswordHash: passwordHash, PhoneNumber: phone}, nil
}

func (m *mockRegistrationDB) CreateProfile(_ context.Context, _ *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, _ *string) (*models.Profile, error) {
	m.createdNickname = nickname
	m.createdFirstName = firstName
	m.createdLastName = lastName
	if m.createProfileErr != nil {
		return nil, m.createProfileErr
	}
	return &models.Profile{ID: uuid.New(), UserID: userID, Nickname: nickname, FirstName: firstName, LastName: lastName}, nil
}

func (m *mockRegistrationDB) DeleteUser(_ context.Context, userID uuid.UUID) error {
	m.deletedUserIDs = append(m.deletedUserIDs, userID)
	return nil
}

func validRegisterBody(t *testing.T) string {
	t.Helper()
	return `{"email":"` + testEmail + `","password":"` + testPassword + `","phone_number":"` + testPhoneRaw + `","nickname":"` + testNickname + `","first_name":"` + testFirstNameRaw + `","last_name":"` + testLastNameRaw + `"}`
}

func TestRegisterUserSuccess(t *testing.T) {
	userID := uuid.New()
	db := &mockRegistrationDB{userID: userID}
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(validRegisterBody(t)))
	rec := httptest.NewRecorder()
	NewHandler(db).RegisterUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if db.createdEmail != testEmail || db.createdNickname != testNickname || db.createdFirstName != testFirstName || db.createdLastName == nil || *db.createdLastName != testLastName {
		t.Fatalf("registration fields were not normalized: %+v", db)
	}
	if db.createdPhone == nil || *db.createdPhone != testPhone {
		t.Fatalf("phone number = %v, want trimmed phone", db.createdPhone)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(db.createdHash), []byte(testPassword)); err != nil {
		t.Fatalf("stored password is not a bcrypt hash of submitted password: %v", err)
	}
	if len(db.deletedUserIDs) != 0 {
		t.Fatalf("unexpected rollback deletes: %v", db.deletedUserIDs)
	}
}

func TestRegisterUserInvalidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandler(&mockRegistrationDB{}).RegisterUser(rec, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("not json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterUserValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing email", body: `{"password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "long email", body: `{"email":"` + strings.Repeat("a", longEmailPrefix) + `@x.co","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "invalid email", body: `{"email":"bad","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "missing password", body: `{"email":"` + testEmail + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "short password", body: `{"email":"` + testEmail + `","password":"` + shortPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "long password", body: `{"email":"` + testEmail + `","password":"` + longPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "invalid phone", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","phone_number":"` + invalidPhone + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "missing nickname", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","first_name":"` + testFirstName + `"}`},
		{name: "short nickname", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + shortNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "long nickname", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + longNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "invalid nickname", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + invalidNickname + `","first_name":"` + testFirstName + `"}`},
		{name: "missing first name", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":" "}`},
		{name: "first name contains digits", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"Ann2"}`},
		{name: "long first name", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"` + longName + `"}`},
		{name: "short last name", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `","last_name":"X"}`},
		{name: "long last name", body: `{"email":"` + testEmail + `","password":"` + testPassword + `","nickname":"` + testNickname + `","first_name":"` + testFirstName + `","last_name":"` + longName + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockRegistrationDB{}
			rec := httptest.NewRecorder()
			NewHandler(db).RegisterUser(rec, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if db.emailCheckCalled || db.createdEmail != "" {
				t.Fatalf("database called for invalid request: %+v", db)
			}
		})
	}
}

func TestRegisterUserDatabaseFailures(t *testing.T) {
	commonErr := errors.New(databaseErrorMsg)
	tests := []struct {
		name         string
		db           *mockRegistrationDB
		wantStatus   int
		wantRollback bool
	}{
		{name: "email check error", db: &mockRegistrationDB{checkEmailErr: commonErr}, wantStatus: http.StatusInternalServerError},
		{name: "email exists", db: &mockRegistrationDB{emailExists: true}, wantStatus: http.StatusConflict},
		{name: "nickname check error", db: &mockRegistrationDB{checkNickErr: commonErr}, wantStatus: http.StatusInternalServerError},
		{name: "nickname exists", db: &mockRegistrationDB{nickExists: true}, wantStatus: http.StatusConflict},
		{name: "user creation error", db: &mockRegistrationDB{createUserErr: commonErr}, wantStatus: http.StatusInternalServerError},
		{name: "profile creation error rolls back user", db: &mockRegistrationDB{userID: uuid.New(), createProfileErr: commonErr}, wantStatus: http.StatusInternalServerError, wantRollback: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewHandler(tt.db).RegisterUser(rec, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(validRegisterBody(t))))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if gotRollback := len(tt.db.deletedUserIDs) == 1; gotRollback != tt.wantRollback {
				t.Fatalf("rollback = %v, want %v (deleted ids: %v)", gotRollback, tt.wantRollback, tt.db.deletedUserIDs)
			}
		})
	}
}
