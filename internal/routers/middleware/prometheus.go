package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus возвращает middleware для сбора метрик количества и длительности запросов.
func Prometheus(counter *prometheus.CounterVec, duration *prometheus.HistogramVec) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		counter.WithLabelValues(method, path, status).Inc()
		duration.WithLabelValues(method, path, status).Observe(time.Since(start).Seconds())
	}
}

// NewHTTPRequestsTotal создаёт Prometheus-счётчик общего количества HTTP-запросов.
func NewHTTPRequestsTotal() *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	prometheus.MustRegister(counter)
	return counter
}

// NewHTTPRequestDuration создаёт Prometheus-гистограмму длительности HTTP-запросов.
func NewHTTPRequestDuration() *prometheus.HistogramVec {
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	prometheus.MustRegister(duration)
	return duration
}
