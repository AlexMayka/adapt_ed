package auth

import (
	authPkg "backend/internal/auth"
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"context"
	"errors"
	"testing"
	"time"

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

// ── mockUserRepository ──────────────────────────────────────────────────────

type mockUserRepository struct {
	user               *dto.User
	err                error
	createUser         *dto.User
	createErr          error
	incrementVersion   int
	incrementErr       error
	incrementCalled    bool
}

func (m *mockUserRepository) GetUserByEmail(_ context.Context, _ string) (*dto.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) CreateUser(_ context.Context, user *dto.User) (*dto.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.createUser != nil {
		return m.createUser, nil
	}
	return user, nil
}

func (m *mockUserRepository) GetUserByID(_ context.Context, _ uuid.UUID) (*dto.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) IncrementSessionVersion(_ context.Context, _ uuid.UUID) (int, error) {
	m.incrementCalled = true
	return m.incrementVersion, m.incrementErr
}

// ── mockTokenRepository ─────────────────────────────────────────────────────

type mockTokenRepository struct {
	setErr          error
	hasActive       bool
	hasActiveErr    error
	revokeResult    bool
	revokeErr       error
	revokeAllErr    error
	revokeCalled    bool
	revokeAllCalled bool
}

func (m *mockTokenRepository) SetTokenByUser(_ context.Context, _ uuid.UUID, _ string, _ string, _ time.Time) error {
	return m.setErr
}

func (m *mockTokenRepository) HasActiveToken(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return m.hasActive, m.hasActiveErr
}

func (m *mockTokenRepository) RevokeTokenByUser(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	m.revokeCalled = true
	return m.revokeResult, m.revokeErr
}

func (m *mockTokenRepository) RevokeAllByUser(_ context.Context, _ uuid.UUID) error {
	m.revokeAllCalled = true
	return m.revokeAllErr
}

// ── mockAuthManager ─────────────────────────────────────────────────────────

type mockAuthManager struct {
	accessToken    string
	accessTokenErr error
	refreshToken   string
	refreshExp     time.Time
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

func (m *mockAuthManager) GenerateAccessToken(_ uuid.UUID, _ *uuid.UUID, _ int, _ dto.UserRole) (string, error) {
	return m.accessToken, m.accessTokenErr
}

func (m *mockAuthManager) ParseAccessToken(_ string) (*authPkg.AccessToken, error) {
	return nil, nil
}

func (m *mockAuthManager) GenerateRefreshToken() (string, time.Time) {
	return m.refreshToken, m.refreshExp
}

func (m *mockAuthManager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *mockAuthManager) RefreshTTL() time.Duration { return m.refreshTTL }

// ── mockSessionCache ────────────────────────────────────────────────────────

type mockSessionCache struct {
	setVersionErr  error
	setHashErr     error
	delErr         error
	versionCalled  bool
	delCalled      bool
}

func (m *mockSessionCache) SetSessionVersion(_ context.Context, _ uuid.UUID, _ int, _ time.Duration) error {
	m.versionCalled = true
	return m.setVersionErr
}

func (m *mockSessionCache) SetRefreshTokenHash(_ context.Context, _ uuid.UUID, _ string, _ time.Duration) error {
	return m.setHashErr
}

func (m *mockSessionCache) DelSession(_ context.Context, _ uuid.UUID) error {
	m.delCalled = true
	return m.delErr
}

// ── Хелперы ─────────────────────────────────────────────────────────────────

func hashPassword(password string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h)
}

func testUser() *dto.User {
	now := time.Now()
	hash := hashPassword("Passw0rd!")
	return &dto.User{
		ID:             uuid.New(),
		Role:           dto.RoleStudent,
		Email:          "test@example.com",
		PasswordHash:   &hash,
		LastName:       "Иванов",
		FirstName:      "Иван",
		SessionVersion: 1,
		IsActive:       true,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}
}

func defaultAuthManager() *mockAuthManager {
	return &mockAuthManager{
		accessToken:  "access-token",
		refreshToken: "refresh-token",
		refreshExp:   time.Now().Add(30 * 24 * time.Hour),
		accessTTL:    15 * time.Minute,
		refreshTTL:   30 * 24 * time.Hour,
	}
}

func newTestService(
	userRep *mockUserRepository,
	tokenRep *mockTokenRepository,
	manager *mockAuthManager,
	cache *mockSessionCache,
) *AuthService {
	return NewAuthService(&mockLogger{}, userRep, tokenRep, manager, cache)
}

// ── Login ───────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	user := testUser()
	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	u, tokens, err := svc.Login(context.Background(), "test@example.com", "Passw0rd!", "Mozilla", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}
	if u.ID != user.ID {
		t.Fatalf("Login() userID = %v, want %v", u.ID, user.ID)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("Login() returned empty tokens")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{err: appErr.ErrUserNotFound},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, _, err := svc.Login(context.Background(), "no@user.com", "Passw0rd!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Login() expected error for missing user")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Login() error type = %T, want *AppError", err)
	}
	if ae.Status != 404 {
		t.Fatalf("Login() status = %d, want 404", ae.Status)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	user := testUser()
	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, _, err := svc.Login(context.Background(), "test@example.com", "WrongPass1!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Login() expected error for wrong password")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Login() error type = %T, want *AppError", err)
	}
	if ae.Status != 401 {
		t.Fatalf("Login() status = %d, want 401", ae.Status)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	user := testUser()
	user.IsActive = false

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, _, err := svc.Login(context.Background(), "test@example.com", "Passw0rd!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Login() expected error for inactive user")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Login() error type = %T, want *AppError", err)
	}
	if ae.Status != 403 {
		t.Fatalf("Login() status = %d, want 403", ae.Status)
	}
}

func TestLogin_NilPasswordHash(t *testing.T) {
	user := testUser()
	user.PasswordHash = nil

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, _, err := svc.Login(context.Background(), "test@example.com", "Passw0rd!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Login() expected error for nil password hash")
	}
}

// ── Registration ────────────────────────────────────────────────────────────

func TestRegistration_Success(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	user := &dto.User{
		Email:     "new@user.com",
		LastName:  "Петров",
		FirstName: "Пётр",
	}

	u, tokens, err := svc.Registration(context.Background(), user, "Passw0rd!", "Mozilla", "127.0.0.1")
	if err != nil {
		t.Fatalf("Registration() error: %v", err)
	}
	if u.Email != "new@user.com" {
		t.Fatalf("Registration() email = %v, want new@user.com", u.Email)
	}
	if u.Role != dto.RoleStudent {
		t.Fatalf("Registration() role = %v, want student", u.Role)
	}
	if !u.IsActive {
		t.Fatal("Registration() user should be active")
	}
	if tokens.AccessToken == "" {
		t.Fatal("Registration() returned empty access token")
	}
}

func TestRegistration_EmailConflict(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{createErr: appErr.ErrEmailAlreadyExists},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	user := &dto.User{Email: "dup@user.com", LastName: "Dup", FirstName: "Dup"}

	_, _, err := svc.Registration(context.Background(), user, "Passw0rd!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Registration() expected error for duplicate email")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Registration() error type = %T, want *AppError", err)
	}
	if ae.Status != 409 {
		t.Fatalf("Registration() status = %d, want 409", ae.Status)
	}
}

// ── GetMe ───────────────────────────────────────────────────────────────────

func TestGetMe_Success(t *testing.T) {
	user := testUser()
	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	u, err := svc.GetMe(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetMe() error: %v", err)
	}
	if u.ID != user.ID {
		t.Fatalf("GetMe() userID = %v, want %v", u.ID, user.ID)
	}
}

func TestGetMe_NotFound(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{err: appErr.ErrUserNotFound},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, err := svc.GetMe(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetMe() expected error for missing user")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("GetMe() error type = %T, want *AppError", err)
	}
	if ae.Status != 404 {
		t.Fatalf("GetMe() status = %d, want 404", ae.Status)
	}
}

func TestGetMe_InactiveUser(t *testing.T) {
	user := testUser()
	user.IsActive = false

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, err := svc.GetMe(context.Background(), user.ID)
	if err == nil {
		t.Fatal("GetMe() expected error for inactive user")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("GetMe() error type = %T, want *AppError", err)
	}
	if ae.Status != 403 {
		t.Fatalf("GetMe() status = %d, want 403", ae.Status)
	}
}

// ── Refresh ─────────────────────────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	user := testUser()

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{revokeResult: true},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	tokens, err := svc.Refresh(context.Background(), user.ID, "raw-refresh-token", "Mozilla", "127.0.0.1")
	if err != nil {
		t.Fatalf("Refresh() error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("Refresh() returned empty tokens")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	user := testUser()

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{revokeResult: false},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, err := svc.Refresh(context.Background(), user.ID, "wrong-token", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Refresh() expected error for invalid token")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Refresh() error type = %T, want *AppError", err)
	}
	if ae.Status != 401 {
		t.Fatalf("Refresh() status = %d, want 401", ae.Status)
	}
}

func TestRefresh_UserNotFound(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{err: appErr.ErrUserNotFound},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, err := svc.Refresh(context.Background(), uuid.New(), "token", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Refresh() expected error for missing user")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Refresh() error type = %T, want *AppError", err)
	}
	if ae.Status != 401 {
		t.Fatalf("Refresh() status = %d, want 401", ae.Status)
	}
}

func TestRefresh_InactiveUser(t *testing.T) {
	user := testUser()
	user.IsActive = false

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, err := svc.Refresh(context.Background(), user.ID, "token", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Refresh() expected error for inactive user")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Refresh() error type = %T, want *AppError", err)
	}
	if ae.Status != 403 {
		t.Fatalf("Refresh() status = %d, want 403", ae.Status)
	}
}

// ── Logout ──────────────────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	tokenRep := &mockTokenRepository{revokeResult: true}

	svc := newTestService(
		&mockUserRepository{},
		tokenRep,
		defaultAuthManager(),
		&mockSessionCache{},
	)

	err := svc.Logout(context.Background(), uuid.New(), "raw-refresh-token")
	if err != nil {
		t.Fatalf("Logout() error: %v", err)
	}
	if !tokenRep.revokeCalled {
		t.Fatal("Logout() did not revoke the token")
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{},
		&mockTokenRepository{revokeResult: false},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	err := svc.Logout(context.Background(), uuid.New(), "wrong-token")
	if err == nil {
		t.Fatal("Logout() expected error for invalid token")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("Logout() error type = %T, want *AppError", err)
	}
	if ae.Status != 401 {
		t.Fatalf("Logout() status = %d, want 401", ae.Status)
	}
}

// ── LogoutAll ───────────────────────────────────────────────────────────────

func TestLogoutAll_Success(t *testing.T) {
	tokenRep := &mockTokenRepository{}
	userRep := &mockUserRepository{incrementVersion: 2}
	cache := &mockSessionCache{}

	svc := newTestService(userRep, tokenRep, defaultAuthManager(), cache)

	err := svc.LogoutAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("LogoutAll() error: %v", err)
	}
	if !tokenRep.revokeAllCalled {
		t.Fatal("LogoutAll() did not revoke all tokens")
	}
	if !userRep.incrementCalled {
		t.Fatal("LogoutAll() did not increment session version")
	}
	if !cache.versionCalled {
		t.Fatal("LogoutAll() did not update cache version")
	}
	if !cache.delCalled {
		t.Fatal("LogoutAll() did not clear session cache")
	}
}

func TestLogoutAll_RevokeError(t *testing.T) {
	tokenRep := &mockTokenRepository{revokeAllErr: errors.New("db error")}
	svc := newTestService(&mockUserRepository{}, tokenRep, defaultAuthManager(), &mockSessionCache{})

	err := svc.LogoutAll(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("LogoutAll() expected error when revoke fails")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("LogoutAll() error type = %T, want *AppError", err)
	}
	if ae.Status != 500 {
		t.Fatalf("LogoutAll() status = %d, want 500", ae.Status)
	}
}

func TestLogoutAll_IncrementVersionError(t *testing.T) {
	userRep := &mockUserRepository{incrementErr: errors.New("db error")}
	svc := newTestService(userRep, &mockTokenRepository{}, defaultAuthManager(), &mockSessionCache{})

	err := svc.LogoutAll(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("LogoutAll() expected error when increment fails")
	}
}

// ── RegistrationByAdmin ─────────────────────────────────────────────────────

func TestRegistrationByAdmin_Success(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	user := &dto.User{
		Email:     "admin-created@user.com",
		LastName:  "Сидоров",
		FirstName: "Сидор",
		Role:      dto.RoleTeacher,
	}

	u, password, err := svc.RegistrationByAdmin(context.Background(), user)
	if err != nil {
		t.Fatalf("RegistrationByAdmin() error: %v", err)
	}
	if u.Email != "admin-created@user.com" {
		t.Fatalf("RegistrationByAdmin() email = %v, want admin-created@user.com", u.Email)
	}
	if u.Role != dto.RoleTeacher {
		t.Fatalf("RegistrationByAdmin() role = %v, want teacher", u.Role)
	}
	if len(password) != 16 {
		t.Fatalf("RegistrationByAdmin() password length = %d, want 16", len(password))
	}
	if !u.IsActive {
		t.Fatal("RegistrationByAdmin() user should be active")
	}
}

func TestRegistrationByAdmin_EmailConflict(t *testing.T) {
	svc := newTestService(
		&mockUserRepository{createErr: appErr.ErrEmailAlreadyExists},
		&mockTokenRepository{},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	user := &dto.User{Email: "dup@user.com", LastName: "Dup", FirstName: "Dup", Role: dto.RoleStudent}

	_, _, err := svc.RegistrationByAdmin(context.Background(), user)
	if err == nil {
		t.Fatal("RegistrationByAdmin() expected error for duplicate email")
	}
	var ae *appErr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("RegistrationByAdmin() error type = %T, want *AppError", err)
	}
	if ae.Status != 409 {
		t.Fatalf("RegistrationByAdmin() status = %d, want 409", ae.Status)
	}
}

// ── issueTokens (косвенно через Login) ──────────────────────────────────────

func TestLogin_TokenGenerationError(t *testing.T) {
	user := testUser()
	mgr := defaultAuthManager()
	mgr.accessTokenErr = errors.New("token gen error")

	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{},
		mgr,
		&mockSessionCache{},
	)

	_, _, err := svc.Login(context.Background(), "test@example.com", "Passw0rd!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Login() expected error when token generation fails")
	}
}

func TestLogin_SaveTokenError(t *testing.T) {
	user := testUser()
	svc := newTestService(
		&mockUserRepository{user: user},
		&mockTokenRepository{setErr: errors.New("db error")},
		defaultAuthManager(),
		&mockSessionCache{},
	)

	_, _, err := svc.Login(context.Background(), "test@example.com", "Passw0rd!", "Mozilla", "127.0.0.1")
	if err == nil {
		t.Fatal("Login() expected error when token save fails")
	}
}
