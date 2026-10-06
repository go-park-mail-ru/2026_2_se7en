package handlers

import (
	"context"
	"encoding/json"
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

func validRegisterData() map[string]any {
	return map[string]any{
		"email":        testEmail,
		"password":     testPassword,
		"phone_number": testPhoneRaw,
		"nickname":     testNickname,
		"first_name":   testFirstNameRaw,
		"last_name":    testLastNameRaw,
	}
}

func modifyRegisterData(modifier func(map[string]any)) string {
	data := validRegisterData()
	modifier(data)

	body, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal test data: " + err.Error())
	}

	return string(body)
}

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
	body, err := json.Marshal(validRegisterData())
	if err != nil {
		t.Fatalf("failed to marshal valid register data: %v", err)
	}
	return string(body)
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
		name     string
		modifier func(map[string]any)
	}{
		{name: "missing email", modifier: func(d map[string]any) { delete(d, "email") }},
		{name: "long email", modifier: func(d map[string]any) { d["email"] = strings.Repeat("a", longEmailPrefix) + "@x.co" }},
		{name: "invalid email", modifier: func(d map[string]any) { d["email"] = "bad" }},
		{name: "missing password", modifier: func(d map[string]any) { delete(d, "password") }},
		{name: "short password", modifier: func(d map[string]any) { d["password"] = shortPassword }},
		{name: "long password", modifier: func(d map[string]any) { d["password"] = longPassword }},
		{name: "invalid phone", modifier: func(d map[string]any) { d["phone_number"] = invalidPhone }},
		{name: "missing nickname", modifier: func(d map[string]any) { delete(d, "nickname") }},
		{name: "short nickname", modifier: func(d map[string]any) { d["nickname"] = shortNickname }},
		{name: "long nickname", modifier: func(d map[string]any) { d["nickname"] = longNickname }},
		{name: "invalid nickname", modifier: func(d map[string]any) { d["nickname"] = invalidNickname }},
		{name: "missing first name", modifier: func(d map[string]any) { d["first_name"] = " " }},
		{name: "first name contains digits", modifier: func(d map[string]any) { d["first_name"] = "Ann2" }},
		{name: "long first name", modifier: func(d map[string]any) { d["first_name"] = longName }},
		{name: "short last name", modifier: func(d map[string]any) { d["last_name"] = "X" }},
		{name: "long last name", modifier: func(d map[string]any) { d["last_name"] = longName }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockRegistrationDB{}
			rec := httptest.NewRecorder()
			body := modifyRegisterData(tt.modifier)
			NewHandler(db).RegisterUser(rec, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body)))

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
