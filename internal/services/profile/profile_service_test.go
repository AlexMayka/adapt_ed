package profile

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// ── Моки ────────────────────────────────────────────────────────────────────

type mockLogger struct{}

func (m *mockLogger) Debug(string, ...any)          {}
func (m *mockLogger) Info(string, ...any)            {}
func (m *mockLogger) Warn(string, ...any)            {}
func (m *mockLogger) Error(string, ...any)           {}
func (m *mockLogger) With(...any) logInf.Logger      { return m }
func (m *mockLogger) WithGroup(string) logInf.Logger { return m }

type mockProfileRepo struct {
	profile       *dto.StudentProfile
	err           error
	createErr     error
	deactivateErr error
}

func (m *mockProfileRepo) GetActiveByUserID(_ context.Context, _ uuid.UUID) (*dto.StudentProfile, error) {
	return m.profile, m.err
}

func (m *mockProfileRepo) Create(_ context.Context, p *dto.StudentProfile) (*dto.StudentProfile, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	p.Version = 1
	return p, nil
}

func (m *mockProfileRepo) DeactivateByUserID(_ context.Context, _ uuid.UUID) error {
	return m.deactivateErr
}

func newTestService(repo *mockProfileRepo) *ProfileService {
	return NewProfileService(&mockLogger{}, repo)
}

// ── GetProfile ──────────────────────────────────────────────────────────────

func TestGetProfile_Success(t *testing.T) {
	p := &dto.StudentProfile{ID: uuid.New(), UserID: uuid.New(), DefaultLevel: dto.LevelSimple}
	svc := newTestService(&mockProfileRepo{profile: p})

	got, err := svc.GetProfile(context.Background(), p.UserID)
	if err != nil {
		t.Fatalf("GetProfile() error: %v", err)
	}
	if got.ID != p.ID {
		t.Fatalf("GetProfile() id = %v, want %v", got.ID, p.ID)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := newTestService(&mockProfileRepo{err: appErr.ErrProfileNotFound})

	_, err := svc.GetProfile(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("GetProfile() status = %d, want 404", ae.Status)
	}
}

// ── CreateDefault ───────────────────────────────────────────────────────────

func TestCreateDefault_Success(t *testing.T) {
	svc := newTestService(&mockProfileRepo{})

	p, err := svc.CreateDefault(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("CreateDefault() error: %v", err)
	}
	if p.DefaultLevel != dto.LevelSimple {
		t.Fatalf("CreateDefault() level = %v, want simple", p.DefaultLevel)
	}
	if len(p.Interests) != 0 {
		t.Fatalf("CreateDefault() interests len = %d, want 0", len(p.Interests))
	}
}

func TestCreateDefault_DBError(t *testing.T) {
	svc := newTestService(&mockProfileRepo{createErr: errors.New("db error")})

	_, err := svc.CreateDefault(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 500 {
		t.Fatalf("CreateDefault() status = %d, want 500", ae.Status)
	}
}

// ── UpdateProfile ───────────────────────────────────────────────────────────

func TestUpdateProfile_Success(t *testing.T) {
	existing := &dto.StudentProfile{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		DefaultLevel: dto.LevelSimple,
		Interests:    []uuid.UUID{},
	}
	svc := newTestService(&mockProfileRepo{profile: existing})

	newLevel := dto.LevelMedium
	interests := []uuid.UUID{uuid.New()}

	updated, err := svc.UpdateProfile(context.Background(), existing.UserID, &newLevel, interests)
	if err != nil {
		t.Fatalf("UpdateProfile() error: %v", err)
	}
	if updated.DefaultLevel != dto.LevelMedium {
		t.Fatalf("UpdateProfile() level = %v, want medium", updated.DefaultLevel)
	}
	if len(updated.Interests) != 1 {
		t.Fatalf("UpdateProfile() interests len = %d, want 1", len(updated.Interests))
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	svc := newTestService(&mockProfileRepo{err: appErr.ErrProfileNotFound})

	_, err := svc.UpdateProfile(context.Background(), uuid.New(), nil, nil)
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("UpdateProfile() status = %d, want 404", ae.Status)
	}
}

func TestUpdateProfile_PartialLevel(t *testing.T) {
	existing := &dto.StudentProfile{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		DefaultLevel: dto.LevelSimple,
		Interests:    []uuid.UUID{uuid.New()},
	}
	svc := newTestService(&mockProfileRepo{profile: existing})

	newLevel := dto.LevelAdvanced
	// Обновляем только level, interests остаются
	updated, err := svc.UpdateProfile(context.Background(), existing.UserID, &newLevel, nil)
	if err != nil {
		t.Fatalf("UpdateProfile() error: %v", err)
	}
	if updated.DefaultLevel != dto.LevelAdvanced {
		t.Fatalf("UpdateProfile() level = %v, want advanced", updated.DefaultLevel)
	}
	if len(updated.Interests) != 1 {
		t.Fatalf("UpdateProfile() interests should be preserved, got len = %d", len(updated.Interests))
	}
}
