package middleware

import (
	logInf "backend/internal/logger/interfaces"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns middleware that recovers from panics, logs the stack trace,
// and responds with 500 Internal Server Error.
func Recovery(log logInf.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"error", r,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
