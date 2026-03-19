package middleware

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

// AuthManager описывает проверку access-токена.
type AuthManager interface {
	CheckToken(token string) (*uuid.UUID, *uuid.UUID, int, *dto.UserRole, error)
}

// Authorization проверяет JWT и пробрасывает данные токена в gin.Context.
func Authorization(manager AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtString := c.GetHeader("Authorization")

		if strings.TrimSpace(jwtString) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.ErrCodeUnauthenticated,
				appErr.ErrJWTNullInHeader.Error(),
			))
			return
		}

		if !strings.HasPrefix(jwtString, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.ErrCodeUnauthenticated,
				appErr.ErrJWTInvalid.Error(),
			))
			return
		}

		jwtString = strings.TrimPrefix(jwtString, "Bearer ")

		userID, schoolID, sessionVersion, role, err := manager.CheckToken(jwtString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.ErrCodeUnauthenticated,
				err.Error(),
			))
			return
		}

		if userID == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.ErrCodeUnauthenticated,
				appErr.ErrJWTInvalid.Error(),
			))
			return
		}

		c.Set(dto.CtxUserID, *userID)
		if schoolID != nil {
			c.Set(dto.CtxSchoolID, *schoolID)
		}

		c.Set(dto.CtxSessionVersion, sessionVersion)
		c.Set(dto.CtxRole, *role)

		c.Next()
	}
}
