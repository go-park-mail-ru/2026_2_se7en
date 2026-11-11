package profile

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type testTxManager struct{ err error }

func (m *testTxManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

type testUserRepository struct {
	user                                              *models.User
	err                                               error
	updateEmailErr, updatePhoneErr, updatePasswordErr error
	updatedEmail                                      string
	updatedPhone                                      *string
	updatedPassword                                   string
}

func (r *testUserRepository) FindByID(context.Context, uuid.UUID) (*models.User, error) {
	return r.user, r.err
}
func (r *testUserRepository) UpdateEmail(_ context.Context, _ uuid.UUID, email string) error {
	r.updatedEmail = email
	return r.updateEmailErr
}
func (r *testUserRepository) UpdatePhone(_ context.Context, _ uuid.UUID, phone *string) error {
	r.updatedPhone = phone
	return r.updatePhoneErr
}
func (r *testUserRepository) UpdatePassword(_ context.Context, _ uuid.UUID, hash string) error {
	r.updatedPassword = hash
	return r.updatePasswordErr
}

type testProfileRepository struct {
	profile                           *models.Profile
	err                               error
	updateErr, updateIconErr          error
	updatedID                         uuid.UUID
	updatedNickname, updatedFirstName string
	updatedLastName, updatedBio       *string
	iconID                            uuid.UUID
}

func (r *testProfileRepository) GetWithFileByUserID(context.Context, uuid.UUID) (*models.Profile, error) {
	return r.profile, r.err
}
func (r *testProfileRepository) Update(_ context.Context, id uuid.UUID, nickname, firstName string, lastName, bio *string) error {
	r.updatedID, r.updatedNickname, r.updatedFirstName, r.updatedLastName, r.updatedBio = id, nickname, firstName, lastName, bio
	return r.updateErr
}
func (r *testProfileRepository) UpdateIcon(_ context.Context, _ uuid.UUID, fileID uuid.UUID) error {
	r.iconID = fileID
	return r.updateIconErr
}

type testFileRepository struct {
	file *models.File
	err  error
}

func (r *testFileRepository) Create(context.Context, string, string, string) (*models.File, error) {
	return r.file, r.err
}

func newTestUsecase(ur *testUserRepository, pr *testProfileRepository, fr *testFileRepository) *ProfileUsecaseImpl {
	return NewProfileUsecase(&testTxManager{}, ur, pr, fr, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestGetAndUpdateProfile(t *testing.T) {
	ctx := context.Background()
	userID, profileID := uuid.New(), uuid.New()
	baseUser := &models.User{ID: userID, Email: "old@example.org"}
	baseProfile := &models.Profile{ID: profileID, Nickname: "old", FirstName: "Имя", LastName: ptr("Фамилия"), Bio: ptr("bio")}

	ur, pr := &testUserRepository{user: baseUser}, &testProfileRepository{profile: baseProfile}
	uc := newTestUsecase(ur, pr, &testFileRepository{})
	user, prof, err := uc.GetProfile(ctx, userID)
	if err != nil || user != baseUser || prof != baseProfile {
		t.Fatalf("GetProfile() = %#v, %#v, %v", user, prof, err)
	}

	ur.user = nil
	if _, _, err := uc.GetProfile(ctx, userID); !errors.Is(err, apperrors.ErrUserNotFound) {
		t.Fatalf("missing user error = %v", err)
	}
	ur.user, ur.err = baseUser, errors.New("read user")
	if _, _, err := uc.GetProfile(ctx, userID); !errors.Is(err, ur.err) {
		t.Fatalf("user repository error = %v", err)
	}
	ur.err = nil
	pr.err = errors.New("read profile")
	if _, _, err := uc.GetProfile(ctx, userID); !errors.Is(err, pr.err) {
		t.Fatalf("profile repository error = %v", err)
	}
	pr.err = nil

	email, phone, nickname, firstName, emptyLast, bio := "new@example.org", "+79990000000", "newnick", "Новое имя", "  ", "new bio"
	err = uc.UpdateProfile(ctx, userID, UpdateParams{Email: &email, PhoneNumber: &phone, Nickname: &nickname, FirstName: &firstName, LastName: &emptyLast, Bio: &bio})
	if err != nil {
		t.Fatalf("UpdateProfile() = %v", err)
	}
	if ur.updatedEmail != email || ur.updatedPhone == nil || *ur.updatedPhone != phone || pr.updatedNickname != nickname || pr.updatedFirstName != firstName || pr.updatedLastName != nil || pr.updatedBio != &bio {
		t.Fatalf("profile update fields: user=%#v profile=%#v", ur, pr)
	}

	pr.profile = nil
	if err := uc.UpdateProfile(ctx, userID, UpdateParams{}); !errors.Is(err, apperrors.ErrProfileNotFound) {
		t.Fatalf("missing profile update error = %v", err)
	}
	pr.profile = baseProfile
	ur.user = nil
	if err := uc.UpdateProfile(ctx, userID, UpdateParams{}); !errors.Is(err, apperrors.ErrUserNotFound) {
		t.Fatalf("missing user update error = %v", err)
	}
	ur.user = baseUser
	pr.err = errors.New("profile lookup failed")
	if err := uc.UpdateProfile(ctx, userID, UpdateParams{}); !errors.Is(err, pr.err) {
		t.Fatalf("profile lookup error = %v", err)
	}
	pr.err = nil
	ur.updateEmailErr = errors.New("email update failed")
	if err := uc.UpdateProfile(ctx, userID, UpdateParams{Email: &email}); !errors.Is(err, ur.updateEmailErr) {
		t.Fatalf("email update error = %v", err)
	}
	ur.updateEmailErr = nil
	ur.updatePhoneErr = errors.New("phone update failed")
	if err := uc.UpdateProfile(ctx, userID, UpdateParams{PhoneNumber: &phone}); !errors.Is(err, ur.updatePhoneErr) {
		t.Fatalf("phone update error = %v", err)
	}
	ur.updatePhoneErr = nil
	pr.updateErr = errors.New("profile update failed")
	if err := uc.UpdateProfile(ctx, userID, UpdateParams{}); !errors.Is(err, pr.updateErr) {
		t.Fatalf("profile update error = %v", err)
	}
}

func TestUploadAvatarAndUpdatePassword(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()
	userID, profileID, fileID := uuid.New(), uuid.New(), uuid.New()
	userModel := &models.User{ID: userID, PasswordHash: ""}
	profileModel := &models.Profile{ID: profileID}
	fileModel := &models.File{ID: fileID, URL: "/uploads/avatar.png"}
	ur := &testUserRepository{user: userModel}
	pr := &testProfileRepository{profile: profileModel}
	fr := &testFileRepository{file: fileModel}
	uc := newTestUsecase(ur, pr, fr)

	if _, err := uc.UploadAvatar(ctx, userID, nil); err == nil {
		t.Fatal("nil avatar accepted")
	}
	file := &models.FileUpload{Reader: bytes.NewReader([]byte("image")), Filename: "avatar.png", Size: 5, ContentType: "image/png"}
	gotFile, err := uc.UploadAvatar(ctx, userID, file)
	if err != nil || gotFile != fileModel || pr.iconID != fileID {
		t.Fatalf("UploadAvatar() = %#v, %v", gotFile, err)
	}

	pr.profile = nil
	if _, err := uc.UploadAvatar(ctx, userID, &models.FileUpload{Reader: bytes.NewReader([]byte("x")), Size: 1, ContentType: "image/png"}); !errors.Is(err, apperrors.ErrProfileNotFound) {
		t.Fatalf("missing profile avatar error = %v", err)
	}
	pr.profile = profileModel
	fr.err = errors.New("file insert failed")
	if _, err := uc.UploadAvatar(ctx, userID, &models.FileUpload{Reader: bytes.NewReader([]byte("x")), Size: 1, ContentType: "image/png"}); !errors.Is(err, fr.err) {
		t.Fatalf("file insert error = %v", err)
	}
	fr.err, fr.file = nil, nil
	if _, err := uc.UploadAvatar(ctx, userID, &models.FileUpload{Reader: bytes.NewReader([]byte("x")), Size: 1, ContentType: "image/png"}); !errors.Is(err, apperrors.ErrFileNotCreated) {
		t.Fatalf("nil file error = %v", err)
	}
	fr.file, pr.updateIconErr = fileModel, errors.New("icon link failed")
	if _, err := uc.UploadAvatar(ctx, userID, &models.FileUpload{Reader: bytes.NewReader([]byte("x")), Size: 1, ContentType: "image/png"}); !errors.Is(err, pr.updateIconErr) {
		t.Fatalf("icon update error = %v", err)
	}

	storedHash, err := bcrypt.GenerateFromPassword([]byte("old-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	ur.user.PasswordHash = string(storedHash)
	if err := uc.UpdatePassword(ctx, userID, "wrong", "new-pass"); !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Fatalf("invalid old password = %v", err)
	}
	if err := uc.UpdatePassword(ctx, userID, "old-pass", "new-pass"); err != nil {
		t.Fatalf("UpdatePassword() = %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ur.updatedPassword), []byte("new-pass")); err != nil {
		t.Fatalf("new password hash invalid: %v", err)
	}
	ur.user = nil
	if err := uc.UpdatePassword(ctx, userID, "old-pass", "new-pass"); !errors.Is(err, apperrors.ErrUserNotFound) {
		t.Fatalf("missing user password update = %v", err)
	}
	ur.user = userModel
	txFail := &testTxManager{err: errors.New("transaction failed")}
	if err := NewProfileUsecase(txFail, ur, pr, fr, slog.Default()).UpdatePassword(ctx, userID, "old-pass", "new-pass"); !errors.Is(err, txFail.err) {
		t.Fatalf("transaction error = %v", err)
	}
}

func ptr(value string) *string { return &value }
