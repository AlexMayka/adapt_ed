package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics returns a handler that serves Prometheus metrics.
// @Summary Метрики приложения
// @Description Возвращает метрики приложения в формате Prometheus для внешнего сбора и мониторинга.
// @Tags infra
// @Produce text/plain
// @Success 200 {string} string
// @Router /infra/metrics [get]
func Metrics() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
