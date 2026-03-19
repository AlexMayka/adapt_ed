package auth

import (
	"backend/internal/dto"
	logInf "backend/internal/logger/interfaces"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── Моки ────────────────────────────────────────────────────────────────────

// mockLogger реализует interfaces.Logger для тестов (no-op).
type mockLogger struct{}

func (m *mockLogger) Debug(string, ...any)            {}
func (m *mockLogger) Info(string, ...any)              {}
func (m *mockLogger) Warn(string, ...any)              {}
func (m *mockLogger) Error(string, ...any)             {}
func (m *mockLogger) With(...any) logInf.Logger        { return m }
func (m *mockLogger) WithGroup(string) logInf.Logger   { return m }

// mockSessionsRepo реализует SessionsRepository для тестов.
type mockSessionsRepo struct {
	version    int
	getErr     error
	setCalled  bool
	setVersion int
}

func (m *mockSessionsRepo) GetSessionVersion(_ context.Context, _ uuid.UUID) (int, error) {
	return m.version, m.getErr
}

func (m *mockSessionsRepo) SetSessionVersion(_ context.Context, _ uuid.UUID, version int, _ time.Duration) error {
	m.setCalled = true
	m.setVersion = version
	return nil
}

// mockUserRepo реализует UserRepository для тестов.
type mockUserRepo struct {
	version int
	err     error
}

func (m *mockUserRepo) GetVersionToken(_ context.Context, _ uuid.UUID) (int, error) {
	return m.version, m.err
}

// ── Хелперы ─────────────────────────────────────────────────────────────────

func newTestManager() *Manager {
	return NewAuthManager(
		&mockLogger{},
		"test-secret-key-32bytes!!",
		15*time.Minute,
		30*24*time.Hour,
		&mockSessionsRepo{version: 1},
		&mockUserRepo{version: 1},
	)
}

// ── GenerateAccessToken ─────────────────────────────────────────────────────

func TestGenerateAccessToken_Success(t *testing.T) {
	m := newTestManager()
	uid := uuid.New()
	schoolID := uuid.New()

	token, err := m.GenerateAccessToken(uid, &schoolID, 1, dto.RoleStudent)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}
}

func TestGenerateAccessToken_NilSchoolID(t *testing.T) {
	m := newTestManager()
	uid := uuid.New()

	token, err := m.GenerateAccessToken(uid, nil, 1, dto.RoleTeacher)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}
}

// ── ParseAccessToken ────────────────────────────────────────────────────────

func TestParseAccessToken_Roundtrip(t *testing.T) {
	m := newTestManager()
	uid := uuid.New()
	schoolID := uuid.New()

	token, err := m.GenerateAccessToken(uid, &schoolID, 3, dto.RoleSchoolAdmin)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error: %v", err)
	}

	claims, err := m.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error: %v", err)
	}

	if claims.UserID != uid {
		t.Fatalf("UserID = %v, want %v", claims.UserID, uid)
	}
	if claims.SchoolID != schoolID {
		t.Fatalf("SchoolID = %v, want %v", claims.SchoolID, schoolID)
	}
	if claims.SessionVersion != 3 {
		t.Fatalf("SessionVersion = %d, want 3", claims.SessionVersion)
	}
	if claims.Role != dto.RoleSchoolAdmin {
		t.Fatalf("Role = %v, want %v", claims.Role, dto.RoleSchoolAdmin)
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	log := &mockLogger{}
	sessions := &mockSessionsRepo{version: 1}
	users := &mockUserRepo{version: 1}

	m1 := NewAuthManager(log, "secret-one-32-bytes-long!!!", 15*time.Minute, 30*24*time.Hour, sessions, users)
	m2 := NewAuthManager(log, "secret-two-32-bytes-long!!!", 15*time.Minute, 30*24*time.Hour, sessions, users)

	token, _ := m1.GenerateAccessToken(uuid.New(), nil, 1, dto.RoleStudent)

	_, err := m2.ParseAccessToken(token)
	if err == nil {
		t.Fatal("ParseAccessToken() expected error for wrong secret, got nil")
	}
}

func TestParseAccessToken_InvalidToken(t *testing.T) {
	m := newTestManager()

	_, err := m.ParseAccessToken("not.a.valid.jwt")
	if err == nil {
		t.Fatal("ParseAccessToken() expected error for invalid token, got nil")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	m := NewAuthManager(
		&mockLogger{},
		"test-secret-key-32bytes!!",
		1*time.Millisecond,
		30*24*time.Hour,
		&mockSessionsRepo{version: 1},
		&mockUserRepo{version: 1},
	)
	token, _ := m.GenerateAccessToken(uuid.New(), nil, 1, dto.RoleStudent)

	time.Sleep(10 * time.Millisecond)

	_, err := m.ParseAccessToken(token)
	if err == nil {
		t.Fatal("ParseAccessToken() expected error for expired token, got nil")
	}
}

// ── GenerateRefreshToken ────────────────────────────────────────────────────

func TestGenerateRefreshToken(t *testing.T) {
	m := newTestManager()

	token1, exp1 := m.GenerateRefreshToken()
	token2, _ := m.GenerateRefreshToken()

	if token1 == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}
	if token1 == token2 {
		t.Fatal("GenerateRefreshToken() returned identical tokens")
	}
	if exp1.Before(time.Now()) {
		t.Fatal("GenerateRefreshToken() returned expiration in the past")
	}
}

// ── TTL getters ─────────────────────────────────────────────────────────────

func TestAccessTTL(t *testing.T) {
	m := newTestManager()
	if m.AccessTTL() != 15*time.Minute {
		t.Fatalf("AccessTTL() = %v, want %v", m.AccessTTL(), 15*time.Minute)
	}
}

func TestRefreshTTL(t *testing.T) {
	m := newTestManager()
	if m.RefreshTTL() != 30*24*time.Hour {
		t.Fatalf("RefreshTTL() = %v, want %v", m.RefreshTTL(), 30*24*time.Hour)
	}
}

// ── CheckToken ──────────────────────────────────────────────────────────────

func TestCheckToken_ValidFromCache(t *testing.T) {
	m := NewAuthManager(
		&mockLogger{},
		"test-secret-key-32bytes!!",
		15*time.Minute,
		30*24*time.Hour,
		&mockSessionsRepo{version: 1},
		&mockUserRepo{version: 1},
	)

	token, _ := m.GenerateAccessToken(uuid.New(), nil, 1, dto.RoleStudent)

	userID, _, _, role, err := m.CheckToken(token)
	if err != nil {
		t.Fatalf("CheckToken() error: %v", err)
	}
	if userID == nil {
		t.Fatal("CheckToken() returned nil userID, want non-nil")
	}
	if *role != dto.RoleStudent {
		t.Fatalf("CheckToken() role = %v, want %v", *role, dto.RoleStudent)
	}
}

func TestCheckToken_ValidFromDB_WarmCache(t *testing.T) {
	sessionsRepo := &mockSessionsRepo{version: -1, getErr: errors.New("cache miss")}
	m := NewAuthManager(
		&mockLogger{},
		"test-secret-key-32bytes!!",
		15*time.Minute,
		30*24*time.Hour,
		sessionsRepo,
		&mockUserRepo{version: 1},
	)

	token, _ := m.GenerateAccessToken(uuid.New(), nil, 1, dto.RoleStudent)

	userID, _, _, _, err := m.CheckToken(token)
	if err != nil {
		t.Fatalf("CheckToken() error: %v", err)
	}
	if userID == nil {
		t.Fatal("CheckToken() returned nil userID (fallback to DB should succeed)")
	}

	// Проверяем что кэш был прогрет
	if !sessionsRepo.setCalled {
		t.Fatal("CheckToken() did not warm Redis cache on DB fallback")
	}
	if sessionsRepo.setVersion != 1 {
		t.Fatalf("CheckToken() cached version = %d, want 1", sessionsRepo.setVersion)
	}
}

func TestCheckToken_VersionOutdated(t *testing.T) {
	m := NewAuthManager(
		&mockLogger{},
		"test-secret-key-32bytes!!",
		15*time.Minute,
		30*24*time.Hour,
		&mockSessionsRepo{version: 5},
		&mockUserRepo{version: 5},
	)

	// Токен с version=1, а в кэше/БД version=5 — токен невалиден
	token, _ := m.GenerateAccessToken(uuid.New(), nil, 1, dto.RoleStudent)

	userID, _, _, _, _ := m.CheckToken(token)
	if userID != nil {
		t.Fatal("CheckToken() returned non-nil userID for outdated session version")
	}
}

func TestCheckToken_InvalidJWT(t *testing.T) {
	m := newTestManager()

	userID, _, _, _, err := m.CheckToken("invalid.jwt.token")
	if userID != nil {
		t.Fatal("CheckToken() returned non-nil userID for invalid JWT")
	}
	if err == nil {
		t.Fatal("CheckToken() expected error for invalid JWT")
	}
}

func TestCheckToken_ReturnsTokenData(t *testing.T) {
	uid := uuid.New()
	schoolID := uuid.New()

	m := NewAuthManager(
		&mockLogger{},
		"test-secret-key-32bytes!!",
		15*time.Minute,
		30*24*time.Hour,
		&mockSessionsRepo{version: 2},
		&mockUserRepo{version: 2},
	)

	token, _ := m.GenerateAccessToken(uid, &schoolID, 3, dto.RoleSchoolAdmin)

	userID, retSchoolID, sessionVersion, role, err := m.CheckToken(token)
	if err != nil {
		t.Fatalf("CheckToken() error: %v", err)
	}
	if *userID != uid {
		t.Fatalf("userID = %v, want %v", *userID, uid)
	}
	if *retSchoolID != schoolID {
		t.Fatalf("schoolID = %v, want %v", *retSchoolID, schoolID)
	}
	if sessionVersion != 3 {
		t.Fatalf("sessionVersion = %d, want 3", sessionVersion)
	}
	if *role != dto.RoleSchoolAdmin {
		t.Fatalf("role = %v, want %v", *role, dto.RoleSchoolAdmin)
	}
}
