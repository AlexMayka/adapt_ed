package middleware

import (
	"backend/internal/dto"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequireRole_Allowed(t *testing.T) {
	tests := []struct {
		name    string
		role    dto.UserRole
		allowed []dto.UserRole
	}{
		{"student allowed", dto.RoleStudent, []dto.UserRole{dto.RoleStudent, dto.RoleTeacher}},
		{"teacher allowed", dto.RoleTeacher, []dto.UserRole{dto.RoleTeacher}},
		{"super_admin allowed", dto.RoleSuperAdmin, []dto.UserRole{dto.RoleSchoolAdmin, dto.RoleSuperAdmin}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(func(c *gin.Context) {
				c.Set(dto.CtxRole, tt.role)
				c.Next()
			})
			r.Use(RequireRole(tt.allowed...))
			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, c.Request)

			if w.Code != http.StatusOK {
				t.Fatalf("RequireRole() status = %d, want 200", w.Code)
			}
		})
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(func(c *gin.Context) {
		c.Set(dto.CtxRole, dto.RoleStudent)
		c.Next()
	})
	r.Use(RequireRole(dto.RoleSchoolAdmin, dto.RoleSuperAdmin))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("RequireRole() status = %d, want 403", w.Code)
	}
}

func TestRequireRole_NoRoleInContext(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(RequireRole(dto.RoleStudent))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("RequireRole() status = %d, want 403", w.Code)
	}
}
