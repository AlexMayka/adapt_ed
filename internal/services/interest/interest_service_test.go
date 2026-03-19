package interest

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

type mockInterestRepo struct {
	interest     *dto.Interest
	interests    []*dto.Interest
	total        int
	err          error
	createErr    error
	updateErr    error
	deleteErr    error
	verifyCount  int
	verifyErr    error
}

func (m *mockInterestRepo) GetByID(_ context.Context, _ uuid.UUID) (*dto.Interest, error) {
	return m.interest, m.err
}

func (m *mockInterestRepo) List(_ context.Context, _ dto.InterestFilter) ([]*dto.Interest, int, error) {
	return m.interests, m.total, m.err
}

func (m *mockInterestRepo) Create(_ context.Context, i *dto.Interest) (*dto.Interest, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return i, nil
}

func (m *mockInterestRepo) Update(_ context.Context, i *dto.Interest) (*dto.Interest, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return i, nil
}

func (m *mockInterestRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func (m *mockInterestRepo) VerifyBatch(_ context.Context, _ []uuid.UUID) (int, error) {
	return m.verifyCount, m.verifyErr
}

// ── Хелперы ─────────────────────────────────────────────────────────────────

func newTestService(repo *mockInterestRepo) *InterestService {
	return NewInterestService(&mockLogger{}, repo)
}

func testInterest() *dto.Interest {
	return &dto.Interest{
		ID:         uuid.New(),
		Name:       "Футбол",
		IsVerified: true,
	}
}

// ── GetInterest ─────────────────────────────────────────────────────────────

func TestGetInterest_Success(t *testing.T) {
	i := testInterest()
	svc := newTestService(&mockInterestRepo{interest: i})

	got, err := svc.GetInterest(context.Background(), i.ID)
	if err != nil {
		t.Fatalf("GetInterest() error: %v", err)
	}
	if got.ID != i.ID {
		t.Fatalf("GetInterest() id = %v, want %v", got.ID, i.ID)
	}
}

func TestGetInterest_NotFound(t *testing.T) {
	svc := newTestService(&mockInterestRepo{err: appErr.ErrInterestNotFound})

	_, err := svc.GetInterest(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("GetInterest() status = %d, want 404", ae.Status)
	}
}

// ── ListInterests ───────────────────────────────────────────────────────────

func TestListInterests_Success(t *testing.T) {
	list := []*dto.Interest{testInterest(), testInterest()}
	svc := newTestService(&mockInterestRepo{interests: list, total: 2})

	got, total, err := svc.ListInterests(context.Background(), dto.InterestFilter{Limit: 20})
	if err != nil {
		t.Fatalf("ListInterests() error: %v", err)
	}
	if len(got) != 2 || total != 2 {
		t.Fatalf("ListInterests() count = %d, total = %d", len(got), total)
	}
}

// ── CreateInterest ──────────────────────────────────────────────────────────

func TestCreateInterest_Success(t *testing.T) {
	svc := newTestService(&mockInterestRepo{})

	created, err := svc.CreateInterest(context.Background(), &dto.Interest{Name: "Шахматы"})
	if err != nil {
		t.Fatalf("CreateInterest() error: %v", err)
	}
	if created.Name != "Шахматы" {
		t.Fatalf("CreateInterest() name = %v, want Шахматы", created.Name)
	}
	if !created.IsVerified {
		t.Fatal("CreateInterest() should be verified")
	}
}

func TestCreateInterest_AlreadyExists(t *testing.T) {
	svc := newTestService(&mockInterestRepo{createErr: appErr.ErrInterestAlreadyExists})

	_, err := svc.CreateInterest(context.Background(), &dto.Interest{Name: "Футбол"})
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 409 {
		t.Fatalf("CreateInterest() status = %d, want 409", ae.Status)
	}
}

// ── UpdateInterest ──────────────────────────────────────────────────────────

func TestUpdateInterest_Success(t *testing.T) {
	existing := testInterest()
	svc := newTestService(&mockInterestRepo{interest: existing})

	updated, err := svc.UpdateInterest(context.Background(), &dto.Interest{ID: existing.ID, Name: "Хоккей"})
	if err != nil {
		t.Fatalf("UpdateInterest() error: %v", err)
	}
	if updated.Name != "Хоккей" {
		t.Fatalf("UpdateInterest() name = %v, want Хоккей", updated.Name)
	}
}

func TestUpdateInterest_NotFound(t *testing.T) {
	svc := newTestService(&mockInterestRepo{err: appErr.ErrInterestNotFound})

	_, err := svc.UpdateInterest(context.Background(), &dto.Interest{ID: uuid.New(), Name: "X"})
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("UpdateInterest() status = %d, want 404", ae.Status)
	}
}

// ── DeleteInterest ──────────────────────────────────────────────────────────

func TestDeleteInterest_Success(t *testing.T) {
	svc := newTestService(&mockInterestRepo{})

	err := svc.DeleteInterest(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("DeleteInterest() error: %v", err)
	}
}

func TestDeleteInterest_NotFound(t *testing.T) {
	svc := newTestService(&mockInterestRepo{deleteErr: appErr.ErrInterestNotFound})

	err := svc.DeleteInterest(context.Background(), uuid.New())
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("DeleteInterest() status = %d, want 404", ae.Status)
	}
}

// ── VerifyInterests ─────────────────────────────────────────────────────────

func TestVerifyInterests_Success(t *testing.T) {
	svc := newTestService(&mockInterestRepo{verifyCount: 3})

	count, err := svc.VerifyInterests(context.Background(), []uuid.UUID{uuid.New(), uuid.New(), uuid.New()})
	if err != nil {
		t.Fatalf("VerifyInterests() error: %v", err)
	}
	if count != 3 {
		t.Fatalf("VerifyInterests() count = %d, want 3", count)
	}
}

func TestVerifyInterests_DBError(t *testing.T) {
	svc := newTestService(&mockInterestRepo{verifyErr: errors.New("db error")})

	_, err := svc.VerifyInterests(context.Background(), []uuid.UUID{uuid.New()})
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 500 {
		t.Fatalf("VerifyInterests() status = %d, want 500", ae.Status)
	}
}
