package middleware

import (
	"backend/internal/dto"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ── Мок AuthManager ──────────────────────────────────────────────────────────

type mockAuthManager struct {
	userID   *uuid.UUID
	schoolID *uuid.UUID
	version  int
	role     *dto.UserRole
	err      error
}

func (m *mockAuthManager) CheckToken(_ string) (*uuid.UUID, *uuid.UUID, int, *dto.UserRole, error) {
	return m.userID, m.schoolID, m.version, m.role, m.err
}

// ── Тесты ───────────────────────────────────────────────────────────────────

func TestAuthorization_Success(t *testing.T) {
	uid := uuid.New()
	schoolID := uuid.New()
	role := dto.RoleStudent

	mgr := &mockAuthManager{
		userID:   &uid,
		schoolID: &schoolID,
		version:  1,
		role:     &role,
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	var gotUserID, gotSchoolID, gotRole interface{}
	var gotVersion interface{}

	r.Use(Authorization(mgr))
	r.GET("/test", func(c *gin.Context) {
		gotUserID, _ = c.Get(dto.CtxUserID)
		gotSchoolID, _ = c.Get(dto.CtxSchoolID)
		gotVersion, _ = c.Get(dto.CtxSessionVersion)
		gotRole, _ = c.Get(dto.CtxRole)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Authorization() status = %d, want 200", w.Code)
	}
	if gotUserID.(uuid.UUID) != uid {
		t.Fatalf("CtxUserID = %v, want %v", gotUserID, uid)
	}
	if gotSchoolID.(uuid.UUID) != schoolID {
		t.Fatalf("CtxSchoolID = %v, want %v", gotSchoolID, schoolID)
	}
	if gotVersion.(int) != 1 {
		t.Fatalf("CtxSessionVersion = %v, want 1", gotVersion)
	}
	if gotRole.(dto.UserRole) != dto.RoleStudent {
		t.Fatalf("CtxRole = %v, want student", gotRole)
	}
}

func TestAuthorization_NoHeader(t *testing.T) {
	mgr := &mockAuthManager{}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Authorization(mgr))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Authorization() status = %d, want 401", w.Code)
	}
}

func TestAuthorization_NoBearerPrefix(t *testing.T) {
	mgr := &mockAuthManager{}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Authorization(mgr))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "just-a-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Authorization() status = %d, want 401", w.Code)
	}
}

func TestAuthorization_InvalidToken(t *testing.T) {
	mgr := &mockAuthManager{err: errors.New("invalid token")}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Authorization(mgr))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Authorization() status = %d, want 401", w.Code)
	}
}

func TestAuthorization_NilUserID(t *testing.T) {
	role := dto.RoleStudent
	mgr := &mockAuthManager{
		userID: nil,
		role:   &role,
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Authorization(mgr))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Authorization() status = %d, want 401", w.Code)
	}
}

func TestAuthorization_NilSchoolID(t *testing.T) {
	uid := uuid.New()
	role := dto.RoleStudent

	mgr := &mockAuthManager{
		userID:   &uid,
		schoolID: nil,
		version:  1,
		role:     &role,
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	var hasSchoolID bool

	r.Use(Authorization(mgr))
	r.GET("/test", func(c *gin.Context) {
		_, hasSchoolID = c.Get(dto.CtxSchoolID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Authorization() status = %d, want 200", w.Code)
	}
	if hasSchoolID {
		t.Fatal("Authorization() should not set CtxSchoolID when schoolID is nil")
	}
}
