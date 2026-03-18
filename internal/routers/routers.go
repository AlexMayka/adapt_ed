package routers

import (
	authPkg "backend/internal/auth"
	"backend/internal/config"
	logInf "backend/internal/logger/interfaces"
	repoSessions "backend/internal/repositories/sessions"
	repoTokens "backend/internal/repositories/tokens"
	repoUsers "backend/internal/repositories/users"
	"backend/internal/routers/handlers"
	"backend/internal/routers/handlers/auth"
	"backend/internal/routers/middleware"
	authSvc "backend/internal/services/auth"
	stgInf "backend/internal/storage/interfaces"
	"backend/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Deps содержит внешние зависимости, необходимые роутеру.
type Deps struct {
	Logger logInf.Logger
	Config config.Config
	DB     *postgres.PoolPsg
	Cache  stgInf.CacheStorage
	S3     stgInf.S3Storage
}

// NewRouter создаёт Gin-движок с цепочкой middleware и маршрутами.
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

	// Инфраструктурные эндпоинты
	infra := r.Group("/infra")
	{
		infra.GET("/health", handlers.Health())
		infra.GET("/ready", handlers.Ready(deps.DB, deps.Cache))
		infra.GET("/metrics", handlers.Metrics())
	}

	// Сборка зависимостей авторизации
	userRepo := repoUsers.NewUserRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	tokenRepo := repoTokens.NewTokenRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	sessionCache := repoSessions.NewSessionRepository(deps.Cache)
	authManager := authPkg.NewAuthManager(deps.Logger, deps.Config.App.Secret, deps.Config.Auth.AccessTTL, deps.Config.Auth.RefreshTTL, sessionCache, userRepo)
	authService := authSvc.NewAuthService(deps.Logger, userRepo, tokenRepo, authManager, sessionCache)

	// Auth endpoints (публичные — без middleware авторизации)
	authH := auth.NewAuthHandlers(deps.Logger, authService)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/registration", authH.Registration)
		authGroup.POST("/registration/admin", authH.RegistrationByAdmin)
		authGroup.POST("/login", authH.Login)
		authGroup.GET("/me", authH.GetMe)
		authGroup.POST("/refresh", authH.Refresh)
		authGroup.POST("/logout", authH.Logout)
		authGroup.POST("/logout-all", authH.LogoutAll)
	}

	return r
}
