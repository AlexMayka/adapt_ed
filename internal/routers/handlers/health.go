package handlers

import (
	stgInf "backend/internal/storage/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

type ReadyResponse struct {
	Status string `json:"status" example:"ready"`
}

type ErrorResponse struct {
	Status string `json:"status" example:"unavailable"`
	Reason string `json:"reason" example:"db: connection refused"`
}

// Health returns a liveness probe handler.
// @Summary Проверка доступности сервиса
// @Description Возвращает успешный ответ, если HTTP-сервис запущен и способен обрабатывать запросы.
// @Tags infra
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /infra/health [get]
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
	}
}

// Ready returns a readiness probe handler.
// @Summary Проверка готовности сервиса
// @Description Проверяет готовность приложения к обработке трафика и доступность критичных зависимостей, таких как база данных и кэш.
// @Tags infra
// @Produce json
// @Success 200 {object} ReadyResponse
// @Failure 503 {object} ErrorResponse
// @Router /infra/ready [get]
func Ready(db stgInf.DbStorage, cache stgInf.CacheStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Status: "unavailable",
				Reason: "db: " + err.Error(),
			})
			return
		}

		if err := cache.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Status: "unavailable",
				Reason: "cache: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, ReadyResponse{Status: "ready"})
	}
}
