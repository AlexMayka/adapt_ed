package middleware

import (
	appErr "backend/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AuthManager interface {
	CheckToken(token string) (bool, error)
}

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

		ok, err := manager.CheckToken(jwtString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.ErrCodeUnauthenticated,
				err.Error(),
			))
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.ErrCodeUnauthenticated,
				appErr.ErrJWTInvalid.Error(),
			))
		}

		c.Next()
	}
}
