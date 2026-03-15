package middleware

import (
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func RequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Request.Header.Get("X-Request-Id")
		if id == "" {
			id = utils.GetUniqUUID().String()
			c.Request.Header.Set("X-Request-Id", id)
		}
		c.Header("X-Request-Id", id)
		c.Next()
	}
}
