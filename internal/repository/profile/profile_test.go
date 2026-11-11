package profile

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

var profileColumns = []string{"id", "icon_id", "user_id", "nickname", "first_name", "last_name", "bio", "created_at", "updated_at"}

func TestProfileRepositoryQueries(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewProfileRepository(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()
	userID, profileID, iconID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	lastName, bio := "Иванов", "bio"
	row := func() *pgxmock.Rows {
		return pgxmock.NewRows(profileColumns).AddRow(profileID, &iconID, userID, "nick", "Иван", &lastName, &bio, now, now)
	}

	pool.ExpectQuery("from profile where user_id").WithArgs(userID).WillReturnRows(row())
	got, err := repo.FindByUserID(ctx, userID)
	if err != nil || got.Nickname != "nick" {
		t.Fatalf("FindByUserID() = %#v, %v (%v)", got, err, errors.Unwrap(err))
	}
	pool.ExpectQuery("select exists").WithArgs("nick").WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	exists, err := repo.CheckNicknameExist(ctx, "nick")
	if err != nil || !exists {
		t.Fatalf("CheckNicknameExist() = %t, %v", exists, err)
	}
	pool.ExpectQuery("insert into profile").WithArgs(&iconID, userID, "nick", "Иван", ptrString("Иванов"), ptrString("bio")).WillReturnRows(row())
	created, err := repo.CreateProfile(ctx, &iconID, userID, "nick", "Иван", ptrString("Иванов"), ptrString("bio"))
	if err != nil || created.ID != profileID {
		t.Fatalf("CreateProfile() = %#v, %v", created, err)
	}

	fileColumns := append(append([]string{}, profileColumns...), "url", "filename", "type", "file_created_at", "file_updated_at")
	pool.ExpectQuery("left join file").WithArgs(userID).WillReturnRows(pgxmock.NewRows(fileColumns).AddRow(profileID, &iconID, userID, "nick", "Иван", &lastName, &bio, now, now, ptrString("/avatar.png"), ptrString("avatar.png"), ptrString("image/png"), &now, &now))
	withFile, err := repo.GetWithFileByUserID(ctx, userID)
	if err != nil || withFile.AvatarURL == nil || *withFile.AvatarURL != "/avatar.png" {
		t.Fatalf("GetWithFileByUserID() = %#v, %v", withFile, err)
	}

	pool.ExpectExec("update profile set nickname").WithArgs(profileID, "new", "Иван", pgxmock.AnyArg(), ptrString("bio")).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.Update(ctx, profileID, "new", "Иван", nil, ptrString("bio")); err != nil {
		t.Fatalf("Update() = %v (%v)", err, errors.Unwrap(err))
	}
	pool.ExpectQuery("join app_user").WithArgs(userID).WillReturnRows(row())
	if got, err := repo.FindProfile(ctx, userID); err != nil || got.ID != profileID {
		t.Fatalf("FindProfile() = %#v, %v", got, err)
	}
	pool.ExpectQuery("update profile").WithArgs(userID, "new", "Иван", pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnRows(row())
	if got, err := repo.UpdateProfile(ctx, userID, models.Profile{Nickname: "new", FirstName: "Иван"}); err != nil || got.ID != profileID {
		t.Fatalf("UpdateProfile() = %#v, %v", got, err)
	}
	pool.ExpectQuery("select password_hash").WithArgs(userID).WillReturnRows(pgxmock.NewRows([]string{"password_hash"}).AddRow("hashed"))
	if hash, err := repo.FindPasswordHash(ctx, userID); err != nil || hash != "hashed" {
		t.Fatalf("FindPasswordHash() = %q, %v", hash, err)
	}
	pool.ExpectExec("update app_user set password_hash").WithArgs(userID, "old", "new").WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if ok, err := repo.UpdatePassword(ctx, userID, "old", "new"); err != nil || !ok {
		t.Fatalf("UpdatePassword() = %t, %v", ok, err)
	}
	pool.ExpectExec("update app_user set phone_number").WithArgs(userID, pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.UpdatePhone(ctx, userID, ptrString("phone")); err != nil {
		t.Fatal(err)
	}
	pool.ExpectExec("update profile set icon_id").WithArgs(profileID, iconID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.UpdateIcon(ctx, profileID, iconID); err != nil {
		t.Fatal(err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProfileRepositoryNotFoundAndErrors(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewProfileRepository(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()
	userID, profileID := uuid.New(), uuid.New()
	pool.ExpectQuery("from profile where user_id").WithArgs(userID).WillReturnError(pgx.ErrNoRows)
	if got, err := repo.FindByUserID(ctx, userID); err != nil || got != nil {
		t.Fatalf("missing profile = %#v, %v", got, err)
	}
	pool.ExpectQuery("left join file").WithArgs(userID).WillReturnError(pgx.ErrNoRows)
	if got, err := repo.GetWithFileByUserID(ctx, userID); err != nil || got != nil {
		t.Fatalf("missing profile/file = %#v, %v", got, err)
	}
	pool.ExpectQuery("join app_user").WithArgs(userID).WillReturnError(pgx.ErrNoRows)
	if got, err := repo.FindProfile(ctx, userID); err != nil || got != nil {
		t.Fatalf("missing active profile = %#v, %v", got, err)
	}
	pool.ExpectQuery("update profile").WithArgs(userID, "n", "f", pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(pgx.ErrNoRows)
	if got, err := repo.UpdateProfile(ctx, userID, models.Profile{Nickname: "n", FirstName: "f"}); err != nil || got != nil {
		t.Fatalf("no matching profile update = %#v, %v (%v)", got, err, errors.Unwrap(err))
	}
	pool.ExpectExec("update profile set nickname").WithArgs(profileID, "n", "f", pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(&pgconn.PgError{Code: "23505"})
	if err := repo.Update(ctx, profileID, "n", "f", nil, nil); err != apperrors.ErrNicknameTaken {
		t.Fatalf("duplicate nickname error = %v (%v)", err, errors.Unwrap(err))
	}
	pool.ExpectExec("update app_user set password_hash").WithArgs(userID, "old", "new").WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	if ok, err := repo.UpdatePassword(ctx, userID, "old", "new"); err != nil || ok {
		t.Fatalf("no-op password update = %t, %v", ok, err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func ptrString(value string) *string { return &value }
