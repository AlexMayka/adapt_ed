package school

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

type mockSchoolRepo struct {
	school      *dto.School
	schools     []*dto.School
	total       int
	err         error
	restoreSchool *dto.School
	createErr     error
	updateErr     error
	restoreErr    error
	deleteErr     error
}

func (m *mockSchoolRepo) GetByID(_ context.Context, _ uuid.UUID) (*dto.School, error) {
	return m.school, m.err
}

func (m *mockSchoolRepo) List(_ context.Context, _ dto.SchoolFilter) ([]*dto.School, int, error) {
	return m.schools, m.total, m.err
}

func (m *mockSchoolRepo) Create(_ context.Context, s *dto.School) (*dto.School, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return s, nil
}

func (m *mockSchoolRepo) Update(_ context.Context, s *dto.School) (*dto.School, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return s, nil
}

func (m *mockSchoolRepo) Restore(_ context.Context, _ uuid.UUID) (*dto.School, error) {
	return m.restoreSchool, m.restoreErr
}

func (m *mockSchoolRepo) SoftDelete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

// ── Хелперы ─────────────────────────────────────────────────────────────────

func newTestService(repo *mockSchoolRepo) *SchoolService {
	return NewSchoolService(&mockLogger{}, repo)
}

func testSchool() *dto.School {
	return &dto.School{
		ID:   uuid.New(),
		Name: "Гимназия №1",
		City: "Москва",
	}
}

// ── GetSchool ───────────────────────────────────────────────────────────────

func TestGetSchool_Success(t *testing.T) {
	s := testSchool()
	svc := newTestService(&mockSchoolRepo{school: s})

	got, err := svc.GetSchool(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("GetSchool() error: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("GetSchool() id = %v, want %v", got.ID, s.ID)
	}
}

func TestGetSchool_NotFound(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{err: appErr.ErrSchoolNotFound})

	_, err := svc.GetSchool(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetSchool() expected error for missing school")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("GetSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 404 {
		t.Fatalf("GetSchool() status = %d, want 404", ae.Status)
	}
}

func TestGetSchool_DBError(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{err: errors.New("db error")})

	_, err := svc.GetSchool(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetSchool() expected error for db failure")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("GetSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 500 {
		t.Fatalf("GetSchool() status = %d, want 500", ae.Status)
	}
}

// ── ListSchools ─────────────────────────────────────────────────────────────

func TestListSchools_Success(t *testing.T) {
	schools := []*dto.School{testSchool(), testSchool()}
	svc := newTestService(&mockSchoolRepo{schools: schools, total: 2})

	got, total, err := svc.ListSchools(context.Background(), dto.SchoolFilter{Limit: 20})
	if err != nil {
		t.Fatalf("ListSchools() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListSchools() count = %d, want 2", len(got))
	}
	if total != 2 {
		t.Fatalf("ListSchools() total = %d, want 2", total)
	}
}

func TestListSchools_DefaultLimit(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{schools: nil, total: 0})

	// limit=0 должен нормализоваться в 20
	_, _, err := svc.ListSchools(context.Background(), dto.SchoolFilter{Limit: 0})
	if err != nil {
		t.Fatalf("ListSchools() error: %v", err)
	}
}

func TestListSchools_NegativeOffset(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{schools: nil, total: 0})

	// offset=-5 должен нормализоваться в 0
	_, _, err := svc.ListSchools(context.Background(), dto.SchoolFilter{Offset: -5})
	if err != nil {
		t.Fatalf("ListSchools() error: %v", err)
	}
}

func TestListSchools_DBError(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{err: errors.New("db error")})

	_, _, err := svc.ListSchools(context.Background(), dto.SchoolFilter{})
	if err == nil {
		t.Fatal("ListSchools() expected error for db failure")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("ListSchools() error type = %T, want *AppError", err)
	}
	if ae.Status != 500 {
		t.Fatalf("ListSchools() status = %d, want 500", ae.Status)
	}
}

// ── CreateSchool ────────────────────────────────────────────────────────────

func TestCreateSchool_Success(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{})

	s := &dto.School{Name: "Новая школа", City: "Казань"}

	created, err := svc.CreateSchool(context.Background(), s)
	if err != nil {
		t.Fatalf("CreateSchool() error: %v", err)
	}
	if created.Name != "Новая школа" {
		t.Fatalf("CreateSchool() name = %v, want 'Новая школа'", created.Name)
	}
	if created.ID == uuid.Nil {
		t.Fatal("CreateSchool() returned nil UUID")
	}
	if created.CreatedAt == nil {
		t.Fatal("CreateSchool() created_at is nil")
	}
}

func TestCreateSchool_DBError(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{createErr: errors.New("db error")})

	_, err := svc.CreateSchool(context.Background(), &dto.School{Name: "Школа", City: "Город"})
	if err == nil {
		t.Fatal("CreateSchool() expected error for db failure")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("CreateSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 500 {
		t.Fatalf("CreateSchool() status = %d, want 500", ae.Status)
	}
}

// ── UpdateSchool ────────────────────────────────────────────────────────────

func TestUpdateSchool_Success(t *testing.T) {
	existing := testSchool()
	svc := newTestService(&mockSchoolRepo{school: existing})

	updated, err := svc.UpdateSchool(context.Background(), &dto.School{
		ID:   existing.ID,
		Name: "Обновлённая",
	})
	if err != nil {
		t.Fatalf("UpdateSchool() error: %v", err)
	}
	if updated.Name != "Обновлённая" {
		t.Fatalf("UpdateSchool() name = %v, want 'Обновлённая'", updated.Name)
	}
	if updated.City != existing.City {
		t.Fatalf("UpdateSchool() city changed to %v, want %v", updated.City, existing.City)
	}
}

func TestUpdateSchool_NotFound(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{err: appErr.ErrSchoolNotFound})

	_, err := svc.UpdateSchool(context.Background(), &dto.School{ID: uuid.New(), Name: "Нет"})
	if err == nil {
		t.Fatal("UpdateSchool() expected error for missing school")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("UpdateSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 404 {
		t.Fatalf("UpdateSchool() status = %d, want 404", ae.Status)
	}
}

func TestUpdateSchool_PartialUpdate(t *testing.T) {
	existing := testSchool()
	existing.City = "Москва"
	svc := newTestService(&mockSchoolRepo{school: existing})

	// Обновляем только город, имя не трогаем
	updated, err := svc.UpdateSchool(context.Background(), &dto.School{
		ID:   existing.ID,
		City: "Воронеж",
	})
	if err != nil {
		t.Fatalf("UpdateSchool() error: %v", err)
	}
	if updated.City != "Воронеж" {
		t.Fatalf("UpdateSchool() city = %v, want 'Воронеж'", updated.City)
	}
	if updated.Name != existing.Name {
		t.Fatalf("UpdateSchool() name changed to %v, want %v", updated.Name, existing.Name)
	}
}

// ── DeleteSchool ────────────────────────────────────────────────────────────

func TestDeleteSchool_Success(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{})

	err := svc.DeleteSchool(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("DeleteSchool() error: %v", err)
	}
}

func TestDeleteSchool_NotFound(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{deleteErr: appErr.ErrSchoolNotFound})

	err := svc.DeleteSchool(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("DeleteSchool() expected error for missing school")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("DeleteSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 404 {
		t.Fatalf("DeleteSchool() status = %d, want 404", ae.Status)
	}
}

func TestDeleteSchool_DBError(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{deleteErr: errors.New("db error")})

	err := svc.DeleteSchool(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("DeleteSchool() expected error for db failure")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("DeleteSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 500 {
		t.Fatalf("DeleteSchool() status = %d, want 500", ae.Status)
	}
}

// ── RestoreSchool ───────────────────────────────────────────────────────────

func TestRestoreSchool_Success(t *testing.T) {
	s := testSchool()
	svc := newTestService(&mockSchoolRepo{restoreSchool: s})

	got, err := svc.RestoreSchool(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("RestoreSchool() error: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("RestoreSchool() id = %v, want %v", got.ID, s.ID)
	}
}

func TestRestoreSchool_NotFound(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{restoreErr: appErr.ErrSchoolNotFound})

	_, err := svc.RestoreSchool(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("RestoreSchool() expected error for missing school")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("RestoreSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 404 {
		t.Fatalf("RestoreSchool() status = %d, want 404", ae.Status)
	}
}

func TestRestoreSchool_DBError(t *testing.T) {
	svc := newTestService(&mockSchoolRepo{restoreErr: errors.New("db error")})

	_, err := svc.RestoreSchool(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("RestoreSchool() expected error for db failure")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("RestoreSchool() error type = %T, want *AppError", err)
	}
	if ae.Status != 500 {
		t.Fatalf("RestoreSchool() status = %d, want 500", ae.Status)
	}
}
