package middleware

import (
	logInf "backend/internal/logger/interfaces"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogging возвращает middleware для структурированного логирования HTTP-запросов.
func StructuredLogging(log logInf.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("http request",
			"method", c.Request.Method,
			"requestId", c.Request.Header.Get("X-Request-Id"),
			"path", c.FullPath(),
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		)
	}
}
