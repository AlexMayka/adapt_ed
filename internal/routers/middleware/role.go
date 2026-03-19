package middleware

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireRole(allowed ...dto.UserRole) gin.HandlerFunc {
	set := make(map[dto.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		set[r] = struct{}{}
	}

	return func(c *gin.Context) {
		val, ok := c.Get(dto.CtxRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, appErr.NewAppError(
				http.StatusForbidden,
				appErr.ErrCodeForbidden,
				"роль пользователя не найдена в контексте",
			))
			return
		}

		role, ok := val.(dto.UserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, appErr.NewAppError(
				http.StatusInternalServerError,
				appErr.ErrCodeInternalServer,
				"некорректный тип роли в контексте",
			))
			return
		}

		if _, exists := set[role]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, appErr.NewAppError(
				http.StatusForbidden,
				appErr.ErrCodeForbidden,
				"недостаточно прав для выполнения операции",
			))
			return
		}

		c.Next()
	}
}
