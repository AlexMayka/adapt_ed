package class

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

type mockClassRepo struct {
	class        *dto.Class
	classes      []*dto.Class
	total        int
	err          error
	createErr    error
	updateErr    error
	deleteErr    error
	restoreClass *dto.Class
	restoreErr   error
}

func (m *mockClassRepo) GetByID(_ context.Context, _ uuid.UUID) (*dto.Class, error) {
	return m.class, m.err
}

func (m *mockClassRepo) List(_ context.Context, _ dto.ClassFilter) ([]*dto.Class, int, error) {
	return m.classes, m.total, m.err
}

func (m *mockClassRepo) Create(_ context.Context, c *dto.Class) (*dto.Class, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return c, nil
}

func (m *mockClassRepo) Update(_ context.Context, c *dto.Class) (*dto.Class, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return c, nil
}

func (m *mockClassRepo) SoftDelete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func (m *mockClassRepo) Restore(_ context.Context, _ uuid.UUID) (*dto.Class, error) {
	return m.restoreClass, m.restoreErr
}

// ── Хелперы ─────────────────────────────────────────────────────────────────

func newTestService(repo *mockClassRepo) *ClassService {
	return NewClassService(&mockLogger{}, repo)
}

func testClass() *dto.Class {
	return &dto.Class{
		ID:              uuid.New(),
		SchoolID:        uuid.New(),
		NumberOfClass:   7,
		SuffixesOfClass: "А",
	}
}

// ── GetClass ────────────────────────────────────────────────────────────────

func TestGetClass_Success(t *testing.T) {
	c := testClass()
	svc := newTestService(&mockClassRepo{class: c})

	got, err := svc.GetClass(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("GetClass() error: %v", err)
	}
	if got.ID != c.ID {
		t.Fatalf("GetClass() id = %v, want %v", got.ID, c.ID)
	}
}

func TestGetClass_NotFound(t *testing.T) {
	svc := newTestService(&mockClassRepo{err: appErr.ErrClassNotFound})

	_, err := svc.GetClass(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetClass() expected error")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("GetClass() status = %d, want 404", ae.Status)
	}
}

// ── ListClasses ─────────────────────────────────────────────────────────────

func TestListClasses_Success(t *testing.T) {
	classes := []*dto.Class{testClass(), testClass()}
	svc := newTestService(&mockClassRepo{classes: classes, total: 2})

	got, total, err := svc.ListClasses(context.Background(), dto.ClassFilter{SchoolID: uuid.New(), Limit: 20})
	if err != nil {
		t.Fatalf("ListClasses() error: %v", err)
	}
	if len(got) != 2 || total != 2 {
		t.Fatalf("ListClasses() count = %d, total = %d", len(got), total)
	}
}

func TestListClasses_DefaultLimit(t *testing.T) {
	svc := newTestService(&mockClassRepo{total: 0})

	_, _, err := svc.ListClasses(context.Background(), dto.ClassFilter{Limit: 0})
	if err != nil {
		t.Fatalf("ListClasses() error: %v", err)
	}
}

// ── CreateClass ─────────────────────────────────────────────────────────────

func TestCreateClass_Success(t *testing.T) {
	svc := newTestService(&mockClassRepo{})

	c := &dto.Class{SchoolID: uuid.New(), NumberOfClass: 7, SuffixesOfClass: "А"}

	created, err := svc.CreateClass(context.Background(), c)
	if err != nil {
		t.Fatalf("CreateClass() error: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("CreateClass() returned nil UUID")
	}
	if created.NumberOfClass != 7 {
		t.Fatalf("CreateClass() number = %d, want 7", created.NumberOfClass)
	}
}

func TestCreateClass_AlreadyExists(t *testing.T) {
	svc := newTestService(&mockClassRepo{createErr: appErr.ErrClassAlreadyExists})

	_, err := svc.CreateClass(context.Background(), &dto.Class{SchoolID: uuid.New(), NumberOfClass: 7, SuffixesOfClass: "А"})
	if err == nil {
		t.Fatal("CreateClass() expected error for duplicate")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 409 {
		t.Fatalf("CreateClass() status = %d, want 409", ae.Status)
	}
}

// ── UpdateClass ─────────────────────────────────────────────────────────────

func TestUpdateClass_Success(t *testing.T) {
	existing := testClass()
	svc := newTestService(&mockClassRepo{class: existing})

	updated, err := svc.UpdateClass(context.Background(), &dto.Class{
		ID:              existing.ID,
		SuffixesOfClass: "Б",
	})
	if err != nil {
		t.Fatalf("UpdateClass() error: %v", err)
	}
	if updated.SuffixesOfClass != "Б" {
		t.Fatalf("UpdateClass() suffix = %v, want Б", updated.SuffixesOfClass)
	}
	if updated.NumberOfClass != existing.NumberOfClass {
		t.Fatalf("UpdateClass() number changed to %d", updated.NumberOfClass)
	}
}

func TestUpdateClass_NotFound(t *testing.T) {
	svc := newTestService(&mockClassRepo{err: appErr.ErrClassNotFound})

	_, err := svc.UpdateClass(context.Background(), &dto.Class{ID: uuid.New(), SuffixesOfClass: "Б"})
	if err == nil {
		t.Fatal("UpdateClass() expected error")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("UpdateClass() status = %d, want 404", ae.Status)
	}
}

func TestUpdateClass_Conflict(t *testing.T) {
	existing := testClass()
	svc := newTestService(&mockClassRepo{class: existing, updateErr: appErr.ErrClassAlreadyExists})

	_, err := svc.UpdateClass(context.Background(), &dto.Class{ID: existing.ID, SuffixesOfClass: "Б"})
	if err == nil {
		t.Fatal("UpdateClass() expected error for conflict")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 409 {
		t.Fatalf("UpdateClass() status = %d, want 409", ae.Status)
	}
}

// ── DeleteClass ─────────────────────────────────────────────────────────────

func TestDeleteClass_Success(t *testing.T) {
	svc := newTestService(&mockClassRepo{})

	err := svc.DeleteClass(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("DeleteClass() error: %v", err)
	}
}

func TestDeleteClass_NotFound(t *testing.T) {
	svc := newTestService(&mockClassRepo{deleteErr: appErr.ErrClassNotFound})

	err := svc.DeleteClass(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("DeleteClass() expected error")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("DeleteClass() status = %d, want 404", ae.Status)
	}
}

// ── RestoreClass ────────────────────────────────────────────────────────────

func TestRestoreClass_Success(t *testing.T) {
	c := testClass()
	svc := newTestService(&mockClassRepo{restoreClass: c})

	got, err := svc.RestoreClass(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("RestoreClass() error: %v", err)
	}
	if got.ID != c.ID {
		t.Fatalf("RestoreClass() id = %v, want %v", got.ID, c.ID)
	}
}

func TestRestoreClass_NotFound(t *testing.T) {
	svc := newTestService(&mockClassRepo{restoreErr: appErr.ErrClassNotFound})

	_, err := svc.RestoreClass(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("RestoreClass() expected error")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) || ae.Status != 404 {
		t.Fatalf("RestoreClass() status = %d, want 404", ae.Status)
	}
}
