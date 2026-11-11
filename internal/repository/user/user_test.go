package user

import (
	"context"
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

var userColumns = []string{"id", "email", "password_hash", "phone_number", "created_at", "updated_at", "deleted_at"}
var userProfileColumns = []string{"id", "icon_id", "user_id", "nickname", "first_name", "last_name", "bio", "created_at", "updated_at"}

func TestUserRepositoryQueries(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewUserRepository(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()
	userID, profileID, sessionID := uuid.New(), uuid.New(), uuid.New()
	now, expiry := time.Now(), time.Now().Add(time.Hour)
	phone, lastName, bio := "+79991234567", "Иванов", "about"
	userRow := func() *pgxmock.Rows {
		return pgxmock.NewRows(userColumns).AddRow(userID, "a@example.org", "hash", &phone, now, now, nil)
	}
	profileRow := func() *pgxmock.Rows {
		return pgxmock.NewRows(userProfileColumns).AddRow(profileID, nil, userID, "nick", "Иван", &lastName, &bio, now, now)
	}

	pool.ExpectQuery("where email = ").WithArgs("a@example.org").WillReturnRows(userRow())
	if got, err := repo.FindByEmail(ctx, "a@example.org"); err != nil || got.ID != userID {
		t.Fatalf("FindByEmail() = %#v, %v", got, err)
	}
	pool.ExpectQuery("where id = ").WithArgs(userID).WillReturnRows(userRow())
	if got, err := repo.FindByID(ctx, userID); err != nil || got.ID != userID {
		t.Fatalf("FindByID() = %#v, %v", got, err)
	}
	pool.ExpectQuery("select exists").WithArgs("a@example.org").WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	if exists, err := repo.CheckEmailExist(ctx, "a@example.org"); err != nil || !exists {
		t.Fatalf("CheckEmailExist() = %t, %v", exists, err)
	}
	pool.ExpectQuery("insert into app_user").WithArgs("a@example.org", "hash", pgxmock.AnyArg()).WillReturnRows(userRow())
	if got, err := repo.CreateUser(ctx, "a@example.org", "hash", &phone); err != nil || got.ID != userID {
		t.Fatalf("CreateUser() = %#v, %v", got, err)
	}

	pool.ExpectBegin()
	pool.ExpectQuery("insert into app_user").WithArgs("a@example.org", "hash", pgxmock.AnyArg()).WillReturnRows(userRow())
	pool.ExpectQuery("insert into profile").WithArgs(pgxmock.AnyArg(), userID, "nick", "Иван", pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnRows(profileRow())
	pool.ExpectQuery("insert into session").WithArgs(userID, expiry).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(sessionID))
	pool.ExpectCommit()
	created, err := repo.CreateUserProfileAndSession(ctx, &models.User{Email: "a@example.org", PhoneNumber: &phone, Profile: &models.Profile{Nickname: "nick", FirstName: "Иван", LastName: &lastName, Bio: &bio}}, "hash", expiry)
	if err != nil || created.SessionID != sessionID || created.User.Profile.Nickname != "nick" {
		t.Fatalf("CreateUserProfileAndSession() = %#v, %v", created, err)
	}

	pool.ExpectExec("delete from app_user").WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repo.DeleteUser(ctx, userID); err != nil {
		t.Fatal(err)
	}
	pool.ExpectExec("update app_user set email").WithArgs(userID, "b@example.org").WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.UpdateEmail(ctx, userID, "b@example.org"); err != nil {
		t.Fatal(err)
	}
	pool.ExpectExec("update app_user set phone_number").WithArgs(userID, pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.UpdatePhone(ctx, userID, ptr(" +70000000000 ")); err != nil {
		t.Fatal(err)
	}
	pool.ExpectExec("update app_user set phone_number").WithArgs(userID, pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.UpdatePhone(ctx, userID, ptr("   ")); err != nil {
		t.Fatal(err)
	}
	pool.ExpectExec("update app_user set password_hash").WithArgs(userID, "new-hash").WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.UpdatePassword(ctx, userID, "new-hash"); err != nil {
		t.Fatal(err)
	}

	pool.ExpectQuery("WHERE p.nickname").WithArgs("nick").WillReturnRows(pgxmock.NewRows(append(append([]string{}, userColumns...), userProfileColumns...)).AddRow(userID, "a@example.org", "hash", &phone, now, now, nil, profileID, nil, userID, "nick", "Иван", &lastName, &bio, now, now))
	if got, err := repo.FindByNickname(ctx, "nick"); err != nil || got.Profile.Nickname != "nick" {
		t.Fatalf("FindByNickname() = %#v, %v", got, err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositoryDatabaseErrors(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewUserRepository(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()
	userID := uuid.New()
	pool.ExpectQuery("where email = ").WithArgs("missing").WillReturnRows(pgxmock.NewRows(userColumns))
	if got, err := repo.FindByEmail(ctx, "missing"); err != nil || got != nil {
		t.Fatalf("missing email = %#v, %v", got, err)
	}
	pool.ExpectQuery("where id = ").WithArgs(userID).WillReturnRows(pgxmock.NewRows(userColumns))
	if got, err := repo.FindByID(ctx, userID); err != nil || got != nil {
		t.Fatalf("missing id = %#v, %v", got, err)
	}
	pool.ExpectBegin()
	pool.ExpectQuery("insert into app_user").WithArgs("a@example.org", "hash", pgxmock.AnyArg()).WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "uq_app_user_email_active"})
	if _, err := repo.CreateUserProfileAndSession(ctx, &models.User{Email: "a@example.org", Profile: &models.Profile{}}, "hash", time.Now()); err != apperrors.ErrEmailTaken {
		t.Fatalf("duplicate email = %v", err)
	}
	pool.ExpectQuery("SELECT au.id").WithArgs("missing").WillReturnError(pgx.ErrNoRows)
	if _, err := repo.FindByNickname(ctx, "missing"); err != apperrors.ErrUserNotFound {
		t.Fatalf("missing nickname = %v", err)
	}
	pool.ExpectExec("update app_user set email").WithArgs(userID, "used@example.org").WillReturnError(&pgconn.PgError{Code: "23505"})
	if err := repo.UpdateEmail(ctx, userID, "used@example.org"); err != apperrors.ErrEmailTaken {
		t.Fatalf("duplicate email update = %v", err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func ptr(value string) *string { return &value }
