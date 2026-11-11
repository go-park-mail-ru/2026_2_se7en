package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
	"github.com/google/uuid"
)

type profileUsecaseStub struct {
	user                                      *models.User
	profile                                   *models.Profile
	getErr, updateErr, uploadErr, passwordErr error
	file                                      *models.File
	updateParams                              usecaseprofile.UpdateParams
}

func (s *profileUsecaseStub) GetProfile(context.Context, uuid.UUID) (*models.User, *models.Profile, error) {
	return s.user, s.profile, s.getErr
}
func (s *profileUsecaseStub) UpdateProfile(_ context.Context, _ uuid.UUID, params usecaseprofile.UpdateParams) error {
	s.updateParams = params
	return s.updateErr
}
func (s *profileUsecaseStub) UploadAvatar(context.Context, uuid.UUID, *models.FileUpload) (*models.File, error) {
	return s.file, s.uploadErr
}
func (s *profileUsecaseStub) UpdatePassword(context.Context, uuid.UUID, string, string) error {
	return s.passwordErr
}

func TestProfileGetUpdateAndPasswordHandlers(t *testing.T) {
	userID := uuid.New()
	uc := &profileUsecaseStub{user: &models.User{ID: userID, Email: "a@example.org"}, profile: &models.Profile{Nickname: "nick", FirstName: "Иван"}}
	delivery := NewProfileDelivery(uc, nil)
	request := func() *http.Request {
		return httptest.NewRequest(http.MethodGet, "/profile", nil).WithContext(middleware.NewAuthContext(context.Background(), userID))
	}

	response := httptest.NewRecorder()
	delivery.HandleGetProfile(response, httptest.NewRequest(http.MethodGet, "/profile", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated profile status = %d", response.Code)
	}
	response = httptest.NewRecorder()
	delivery.HandleGetProfile(response, request())
	if response.Code != http.StatusOK {
		t.Fatalf("profile status = %d", response.Code)
	}
	var body ProfileResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Nickname != "nick" {
		t.Fatalf("profile response = %#v, %v", body, err)
	}

	uc.getErr = apperrors.ErrUserNotFound
	response = httptest.NewRecorder()
	delivery.HandleGetProfile(response, request())
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing profile status = %d", response.Code)
	}
	uc.getErr = errors.New("database unavailable")
	response = httptest.NewRecorder()
	delivery.HandleGetProfile(response, request())
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("get failure status = %d", response.Code)
	}
	uc.getErr = nil

	updateReq := &UpdateProfileRequest{Nickname: ptr("newnick"), PhoneNumber: ptr("+79990000000")}
	response = httptest.NewRecorder()
	if err := delivery.HandleUpdateProfile(response, request(), updateReq); err != nil || response.Code != http.StatusOK {
		t.Fatalf("update profile = %v, status %d", err, response.Code)
	}
	if uc.updateParams.Nickname == nil || *uc.updateParams.Nickname != "newnick" || uc.updateParams.PhoneNumber == nil {
		t.Fatalf("update params not forwarded: %#v", uc.updateParams)
	}
	uc.updateErr = apperrors.ErrNicknameTaken
	if err := delivery.HandleUpdateProfile(httptest.NewRecorder(), request(), updateReq); err == nil || err.Status != http.StatusConflict {
		t.Fatalf("duplicate nickname response = %v", err)
	}
	uc.updateErr = apperrors.ErrProfileNotFound
	if err := delivery.HandleUpdateProfile(httptest.NewRecorder(), request(), updateReq); err == nil || err.Status != http.StatusNotFound {
		t.Fatalf("missing profile response = %v", err)
	}
	uc.updateErr = errors.New("write failure")
	if err := delivery.HandleUpdateProfile(httptest.NewRecorder(), request(), updateReq); err == nil || err.Status != http.StatusInternalServerError {
		t.Fatalf("update failure response = %v", err)
	}
	uc.updateErr = nil
	uc.getErr = errors.New("refresh failure")
	if err := delivery.HandleUpdateProfile(httptest.NewRecorder(), request(), updateReq); err == nil || err.Status != http.StatusInternalServerError {
		t.Fatalf("refresh failure response = %v", err)
	}
	uc.getErr = nil

	if err := delivery.HandleUpdatePassword(httptest.NewRecorder(), httptest.NewRequest(http.MethodPatch, "/profile/password", nil), &UpdatePasswordRequest{}); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated password update = %v", err)
	}
	if err := delivery.HandleUpdatePassword(httptest.NewRecorder(), request(), &UpdatePasswordRequest{OldPassword: "bad", NewPassword: "new-password"}); err != nil {
		t.Fatalf("password update = %v", err)
	}
	uc.passwordErr = apperrors.ErrInvalidCredentials
	if err := delivery.HandleUpdatePassword(httptest.NewRecorder(), request(), &UpdatePasswordRequest{}); err == nil || err.Status != http.StatusBadRequest {
		t.Fatalf("invalid old password response = %v", err)
	}
	uc.passwordErr = errors.New("database failure")
	if err := delivery.HandleUpdatePassword(httptest.NewRecorder(), request(), &UpdatePasswordRequest{}); err == nil || err.Status != http.StatusInternalServerError {
		t.Fatalf("password error response = %v", err)
	}
}

func TestPublicProfileAndAvatarHandlers(t *testing.T) {
	userID := uuid.New()
	uc := &profileUsecaseStub{user: &models.User{ID: userID, Email: "private@example.org"}, profile: &models.Profile{Nickname: "public", FirstName: "Name"}, file: &models.File{URL: "/uploads/avatar.png"}}
	delivery := NewProfileDelivery(uc, nil)
	publicReq := func(id string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/users/"+id, nil).WithContext(middleware.NewAuthContext(context.Background(), userID))
		r.SetPathValue("id", id)
		return r
	}
	response := httptest.NewRecorder()
	delivery.HandleGetPublicProfile(response, publicReq(userID.String()))
	if response.Code != http.StatusOK || bytes.Contains(response.Body.Bytes(), []byte("private@example.org")) {
		t.Fatalf("public profile response leaks data or fails: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	delivery.HandleGetPublicProfile(response, httptest.NewRequest(http.MethodGet, "/users/1", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated public profile = %d", response.Code)
	}
	response = httptest.NewRecorder()
	delivery.HandleGetPublicProfile(response, publicReq("bad-id"))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid public profile id = %d", response.Code)
	}
	uc.profile = nil
	response = httptest.NewRecorder()
	delivery.HandleGetPublicProfile(response, publicReq(userID.String()))
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing public profile = %d", response.Code)
	}
	uc.profile = &models.Profile{}
	uc.getErr = errors.New("lookup failed")
	response = httptest.NewRecorder()
	delivery.HandleGetPublicProfile(response, publicReq(userID.String()))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("public profile failure = %d", response.Code)
	}
	uc.getErr = nil

	avatarRequest := func() *http.Request {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "avatar.png")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write([]byte("\x89PNG\r\n\x1a\ncontent"))
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest(http.MethodPut, "/profile/avatar", &body).WithContext(middleware.NewAuthContext(context.Background(), userID))
		r.Header.Set("Content-Type", writer.FormDataContentType())
		return r
	}
	response = httptest.NewRecorder()
	delivery.HandleUploadAvatar(response, httptest.NewRequest(http.MethodPut, "/profile/avatar", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated avatar upload = %d", response.Code)
	}
	response = httptest.NewRecorder()
	delivery.HandleUploadAvatar(response, httptest.NewRequest(http.MethodPut, "/profile/avatar", bytes.NewReader([]byte("invalid"))).WithContext(middleware.NewAuthContext(context.Background(), userID)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid avatar upload = %d", response.Code)
	}
	response = httptest.NewRecorder()
	delivery.HandleUploadAvatar(response, avatarRequest())
	if response.Code != http.StatusOK {
		t.Fatalf("avatar upload = %d: %s", response.Code, response.Body.String())
	}
	uc.uploadErr = apperrors.ErrProfileNotFound
	response = httptest.NewRecorder()
	delivery.HandleUploadAvatar(response, avatarRequest())
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing profile upload = %d", response.Code)
	}
	uc.uploadErr = errors.New("write failed")
	response = httptest.NewRecorder()
	delivery.HandleUploadAvatar(response, avatarRequest())
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("avatar error = %d", response.Code)
	}
}

func ptr(value string) *string { return &value }
