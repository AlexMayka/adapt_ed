package handlers

import (
	stgInf "backend/internal/storage/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health returns a liveness probe handler.
// Always responds 200 if the process is running.
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// Ready returns a readiness probe handler.
// Pings DB and Cache; responds 200 only when both are reachable.
func Ready(db stgInf.DbStorage, cache stgInf.CacheStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"reason": "db: " + err.Error(),
			})
			return
		}

		if err := cache.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"reason": "cache: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
