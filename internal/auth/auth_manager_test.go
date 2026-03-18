package auth

import (
	"backend/internal/dto"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func newTestManager() *Manager {
	return NewAuthManager("test-secret-key-32bytes!!", 15*time.Minute, 30*24*time.Hour)
}

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
	m1 := NewAuthManager("secret-one-32-bytes-long!!!", 15*time.Minute, 30*24*time.Hour)
	m2 := NewAuthManager("secret-two-32-bytes-long!!!", 15*time.Minute, 30*24*time.Hour)

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
	m := NewAuthManager("test-secret-key-32bytes!!", 1*time.Millisecond, 30*24*time.Hour)
	token, _ := m.GenerateAccessToken(uuid.New(), nil, 1, dto.RoleStudent)

	time.Sleep(10 * time.Millisecond)

	_, err := m.ParseAccessToken(token)
	if err == nil {
		t.Fatal("ParseAccessToken() expected error for expired token, got nil")
	}
}

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

func TestCheckRefreshToken(t *testing.T) {
	m := newTestManager()

	token, _ := m.GenerateRefreshToken()

	// Для CheckRefreshToken нужен bcrypt хэш
	hash, err := hashForTest(token)
	if err != nil {
		t.Fatalf("hashForTest() error: %v", err)
	}

	if !m.CheckRefreshToken(token, hash) {
		t.Fatal("CheckRefreshToken() returned false for valid token")
	}
	if m.CheckRefreshToken("wrong-token", hash) {
		t.Fatal("CheckRefreshToken() returned true for wrong token")
	}
}

func hashForTest(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
