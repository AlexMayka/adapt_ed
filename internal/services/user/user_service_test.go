package user

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ── Моки ────────────────────────────────────────────────────────────────────

type mockLogger struct{}

func (m *mockLogger) Debug(string, ...any)          {}
func (m *mockLogger) Info(string, ...any)            {}
func (m *mockLogger) Warn(string, ...any)            {}
func (m *mockLogger) Error(string, ...any)           {}
func (m *mockLogger) With(...any) logInf.Logger      { return m }
func (m *mockLogger) WithGroup(string) logInf.Logger { return m }

type mockUserRepo struct {
	user        *dto.User
	users       []*dto.User
	total       int
	err         error
	updateErr   error
	passErr     error
	activeErr   error
	deleteErr   error
	restoreUser *dto.User
	restoreErr  error
}

func (m *mockUserRepo) GetUserByID(_ context.Context, _ uuid.UUID) (*dto.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) List(_ context.Context, _ dto.UserFilter) ([]*dto.User, int, error) {
	return m.users, m.total, m.err
}

func (m *mockUserRepo) Update(_ context.Context, u *dto.User) (*dto.User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return u, nil
}

func (m *mockUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return m.passErr
}

func (m *mockUserRepo) SetActive(_ context.Context, _ uuid.UUID, _ bool) error {
	return m.activeErr
}

func (m *mockUserRepo) SoftDelete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func (m *mockUserRepo) Restore(_ context.Context, _ uuid.UUID) (*dto.User, error) {
	return m.restoreUser, m.restoreErr
}

// ── Хелперы ─────────────────────────────────────────────────────────────────

func newTestService(repo *mockUserRepo) *UserService {
	return NewUserService(&mockLogger{}, repo)
}

func hashPass(p string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(h)
}

func testUser() *dto.User {
	hash := hashPass("Passw0rd!")
	return &dto.User{
		ID:           uuid.New(),
		Role:         dto.RoleStudent,
		Email:        "test@example.com",
		PasswordHash: &hash,
		LastName:     "Иванов",
		FirstName:    "Иван",
		IsActive:     true,
	}
}

// ── GetUser ─────────────────────────────────────────────────────────────────

func TestGetUser_Success(t *testing.T) {
	u := testUser()
	svc := newTestService(&mockUserRepo{user: u})

	got, err := svc.GetUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("GetUser() id = %v, want %v", got.ID, u.ID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc := newTestService(&mockUserRepo{err: appErr.ErrUserNotFound})

	_, err := svc.GetUser(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("GetUser() status = %d, want 404", ae.Status)
	}
}

// ── ListUsers ───────────────────────────────────────────────────────────────

func TestListUsers_Success(t *testing.T) {
	users := []*dto.User{testUser(), testUser()}
	svc := newTestService(&mockUserRepo{users: users, total: 2})

	got, total, err := svc.ListUsers(context.Background(), dto.UserFilter{Limit: 20})
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if len(got) != 2 || total != 2 {
		t.Fatalf("ListUsers() count = %d, total = %d", len(got), total)
	}
}

// ── UpdateUser ──────────────────────────────────────────────────────────────

func TestUpdateUser_Success(t *testing.T) {
	existing := testUser()
	svc := newTestService(&mockUserRepo{user: existing})

	updated, err := svc.UpdateUser(context.Background(), &dto.User{ID: existing.ID, Email: "new@example.com"})
	if err != nil {
		t.Fatalf("UpdateUser() error: %v", err)
	}
	if updated.Email != "new@example.com" {
		t.Fatalf("UpdateUser() email = %v, want new@example.com", updated.Email)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc := newTestService(&mockUserRepo{err: appErr.ErrUserNotFound})

	_, err := svc.UpdateUser(context.Background(), &dto.User{ID: uuid.New(), Email: "x@x.com"})
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("UpdateUser() status = %d, want 404", ae.Status)
	}
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	existing := testUser()
	svc := newTestService(&mockUserRepo{user: existing, updateErr: appErr.ErrEmailAlreadyExists})

	_, err := svc.UpdateUser(context.Background(), &dto.User{ID: existing.ID, Email: "dup@x.com"})
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 409 {
		t.Fatalf("UpdateUser() status = %d, want 409", ae.Status)
	}
}

// ── ChangePassword ──────────────────────────────────────────────────────────

func TestChangePassword_Success(t *testing.T) {
	u := testUser()
	svc := newTestService(&mockUserRepo{user: u})

	err := svc.ChangePassword(context.Background(), u.ID, "Passw0rd!", "NewPass1!")
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}
}

func TestChangePassword_WrongOld(t *testing.T) {
	u := testUser()
	svc := newTestService(&mockUserRepo{user: u})

	err := svc.ChangePassword(context.Background(), u.ID, "WrongOld1!", "NewPass1!")
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 401 {
		t.Fatalf("ChangePassword() status = %d, want 401", ae.Status)
	}
}

// ── SetActive ───────────────────────────────────────────────────────────────

func TestSetActive_Success(t *testing.T) {
	svc := newTestService(&mockUserRepo{})

	err := svc.SetActive(context.Background(), uuid.New(), false)
	if err != nil {
		t.Fatalf("SetActive() error: %v", err)
	}
}

func TestSetActive_NotFound(t *testing.T) {
	svc := newTestService(&mockUserRepo{activeErr: appErr.ErrUserNotFound})

	err := svc.SetActive(context.Background(), uuid.New(), false)
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("SetActive() status = %d, want 404", ae.Status)
	}
}

// ── DeleteUser ──────────────────────────────────────────────────────────────

func TestDeleteUser_Success(t *testing.T) {
	svc := newTestService(&mockUserRepo{})

	err := svc.DeleteUser(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("DeleteUser() error: %v", err)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc := newTestService(&mockUserRepo{deleteErr: appErr.ErrUserNotFound})

	err := svc.DeleteUser(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("DeleteUser() status = %d, want 404", ae.Status)
	}
}

// ── RestoreUser ─────────────────────────────────────────────────────────────

func TestRestoreUser_Success(t *testing.T) {
	u := testUser()
	svc := newTestService(&mockUserRepo{restoreUser: u})

	got, err := svc.RestoreUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("RestoreUser() error: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("RestoreUser() id = %v, want %v", got.ID, u.ID)
	}
}

func TestRestoreUser_NotFound(t *testing.T) {
	svc := newTestService(&mockUserRepo{restoreErr: appErr.ErrUserNotFound})

	_, err := svc.RestoreUser(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("RestoreUser() status = %d, want 404", ae.Status)
	}
}
