package routers

import (
	logInf "backend/internal/logger/interfaces"
	"backend/internal/routers/handlers"
	"backend/internal/routers/middleware"
	stgInf "backend/internal/storage/interfaces"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Deps groups all external dependencies required by the router.
type Deps struct {
	Logger logInf.Logger
	DB     stgInf.DbStorage
	Cache  stgInf.CacheStorage
	S3     stgInf.S3Storage
}

// NewRouter creates a Gin engine with middleware stack and infra endpoints.
func NewRouter(deps Deps, envType string) *gin.Engine {
	if envType == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	counter := middleware.NewHTTPRequestsTotal()
	duration := middleware.NewHTTPRequestDuration()

	r.Use(middleware.RequestId())
	r.Use(middleware.Recovery(deps.Logger))
	r.Use(middleware.StructuredLogging(deps.Logger))
	r.Use(middleware.Prometheus(counter, duration))
	r.Use(middleware.CORS())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Infra endpoints
	infra := r.Group("/infra")
	{
		infra.GET("/health", handlers.Health())
		infra.GET("/ready", handlers.Ready(deps.DB, deps.Cache))
		infra.GET("/metrics", handlers.Metrics())
	}

	return r
}
