package routers

import (
	authPkg "backend/internal/auth"
	"backend/internal/config"
	"backend/internal/dto"
	logInf "backend/internal/logger/interfaces"
	repoSchools "backend/internal/repositories/schools"
	repoSessions "backend/internal/repositories/sessions"
	repoTokens "backend/internal/repositories/tokens"
	repoUsers "backend/internal/repositories/users"
	"backend/internal/routers/handlers"
	"backend/internal/routers/handlers/auth"
	"backend/internal/routers/handlers/school"
	"backend/internal/routers/middleware"
	authSvc "backend/internal/services/auth"
	schoolSvc "backend/internal/services/school"
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

	authH := auth.NewAuthHandlers(deps.Logger, authService)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/registration", authH.Registration)
		authGroup.POST("/login", authH.Login)
		authGroup.POST("/refresh", authH.Refresh)
	}

	authMidGroup := r.Group("/auth")
	authMidGroup.Use(middleware.Authorization(authManager))
	{
		authMidGroup.GET("/me", authH.GetMe)
		authMidGroup.POST("/logout", authH.Logout)
		authMidGroup.POST("/logout-all", authH.LogoutAll)
	}

	adminGroup := r.Group("/auth")
	adminGroup.Use(middleware.Authorization(authManager))
	adminGroup.Use(middleware.RequireRole(dto.RoleSchoolAdmin, dto.RoleSuperAdmin))
	{
		adminGroup.POST("/registration/admin", authH.RegistrationByAdmin)
	}

	// Сборка зависимостей школ
	schoolRepo := repoSchools.NewSchoolRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	schoolService := schoolSvc.NewSchoolService(deps.Logger, schoolRepo)
	schoolH := school.NewSchoolHandlers(deps.Logger, schoolService)

	schoolGroup := r.Group("/schools")
	schoolGroup.Use(middleware.Authorization(authManager))
	{
		schoolGroup.GET("", schoolH.ListSchools)
		schoolGroup.GET("/:id", schoolH.GetSchool)
	}

	schoolAdminGroup := r.Group("/schools")
	schoolAdminGroup.Use(middleware.Authorization(authManager))
	schoolAdminGroup.Use(middleware.RequireRole(dto.RoleSchoolAdmin, dto.RoleSuperAdmin))
	{
		schoolAdminGroup.PATCH("/:id", schoolH.UpdateSchool)
	}

	schoolSuperGroup := r.Group("/schools")
	schoolSuperGroup.Use(middleware.Authorization(authManager))
	schoolSuperGroup.Use(middleware.RequireRole(dto.RoleSuperAdmin))
	{
		schoolSuperGroup.POST("", schoolH.CreateSchool)
		schoolSuperGroup.DELETE("/:id", schoolH.DeleteSchool)
		schoolSuperGroup.POST("/:id/restore", schoolH.RestoreSchool)
	}

	return r
}
