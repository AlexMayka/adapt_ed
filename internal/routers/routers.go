package routers

import (
	authPkg "backend/internal/auth"
	"backend/internal/config"
	"backend/internal/dto"
	logInf "backend/internal/logger/interfaces"
	repoClasses "backend/internal/repositories/classes"
	repoInterests "backend/internal/repositories/interests"
	repoProfiles "backend/internal/repositories/profiles"
	repoSchools "backend/internal/repositories/schools"
	repoSessions "backend/internal/repositories/sessions"
	repoTokens "backend/internal/repositories/tokens"
	repoUsers "backend/internal/repositories/users"
	"backend/internal/routers/handlers"
	"backend/internal/routers/handlers/auth"
	classH "backend/internal/routers/handlers/class"
	interestH "backend/internal/routers/handlers/interest"
	profileH "backend/internal/routers/handlers/profile"
	"backend/internal/routers/handlers/school"
	userH "backend/internal/routers/handlers/user"
	"backend/internal/routers/middleware"
	authSvc "backend/internal/services/auth"
	classSvc "backend/internal/services/class"
	interestSvc "backend/internal/services/interest"
	profileSvc "backend/internal/services/profile"
	schoolSvc "backend/internal/services/school"
	userSvc "backend/internal/services/user"
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
	profileRepo := repoProfiles.NewProfileRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	profileService := profileSvc.NewProfileService(deps.Logger, profileRepo)
	authService := authSvc.NewAuthService(deps.Logger, userRepo, tokenRepo, authManager, sessionCache, profileService)

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

	// Сборка зависимостей школ и классов
	schoolRepo := repoSchools.NewSchoolRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	schoolService := schoolSvc.NewSchoolService(deps.Logger, schoolRepo)
	schoolH := school.NewSchoolHandlers(deps.Logger, schoolService)

	classRepo := repoClasses.NewClassRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	classService := classSvc.NewClassService(deps.Logger, classRepo)
	classHandlers := classH.NewClassHandlers(deps.Logger, classService)

	schoolGroup := r.Group("/schools")
	schoolGroup.Use(middleware.Authorization(authManager))
	{
		schoolGroup.GET("", schoolH.ListSchools)
		schoolGroup.GET("/:id", schoolH.GetSchool)
		schoolGroup.GET("/:id/classes", classHandlers.ListClasses)
		schoolGroup.GET("/:id/classes/:class_id", classHandlers.GetClass)
	}

	schoolAdminGroup := r.Group("/schools")
	schoolAdminGroup.Use(middleware.Authorization(authManager))
	schoolAdminGroup.Use(middleware.RequireRole(dto.RoleSchoolAdmin, dto.RoleSuperAdmin))
	{
		schoolAdminGroup.PATCH("/:id", schoolH.UpdateSchool)
		schoolAdminGroup.POST("/:id/classes", classHandlers.CreateClass)
		schoolAdminGroup.PATCH("/:id/classes/:class_id", classHandlers.UpdateClass)
		schoolAdminGroup.DELETE("/:id/classes/:class_id", classHandlers.DeleteClass)
	}

	schoolSuperGroup := r.Group("/schools")
	schoolSuperGroup.Use(middleware.Authorization(authManager))
	schoolSuperGroup.Use(middleware.RequireRole(dto.RoleSuperAdmin))
	{
		schoolSuperGroup.POST("", schoolH.CreateSchool)
		schoolSuperGroup.DELETE("/:id", schoolH.DeleteSchool)
		schoolSuperGroup.POST("/:id/restore", schoolH.RestoreSchool)
		schoolSuperGroup.POST("/:id/classes/:class_id/restore", classHandlers.RestoreClass)
	}

	// Сборка зависимостей пользователей
	userService := userSvc.NewUserService(deps.Logger, userRepo)
	userHandlers := userH.NewUserHandlers(deps.Logger, userService)

	// Операции текущего пользователя
	userSelfGroup := r.Group("/users")
	userSelfGroup.Use(middleware.Authorization(authManager))
	{
		userSelfGroup.PATCH("/me", userHandlers.UpdateProfile)
		userSelfGroup.POST("/me/password", userHandlers.ChangePassword)

		profileHandlers := profileH.NewProfileHandlers(deps.Logger, profileService)
		userSelfGroup.GET("/me/profile", profileHandlers.GetMyProfile)
		userSelfGroup.PATCH("/me/profile", profileHandlers.UpdateMyProfile)
	}

	// Чтение пользователей (admin)
	userReadGroup := r.Group("/users")
	userReadGroup.Use(middleware.Authorization(authManager))
	userReadGroup.Use(middleware.RequireRole(dto.RoleSchoolAdmin, dto.RoleSuperAdmin, dto.RoleTeacher))
	{
		userReadGroup.GET("", userHandlers.ListUsers)
		userReadGroup.GET("/:id", userHandlers.GetUser)
	}

	// Управление пользователями (admin)
	userAdminGroup := r.Group("/users")
	userAdminGroup.Use(middleware.Authorization(authManager))
	userAdminGroup.Use(middleware.RequireRole(dto.RoleSchoolAdmin, dto.RoleSuperAdmin))
	{
		userAdminGroup.PATCH("/:id", userHandlers.UpdateUser)
		userAdminGroup.PATCH("/:id/active", userHandlers.SetActive)
	}

	// Удаление/восстановление (super_admin)
	userSuperGroup := r.Group("/users")
	userSuperGroup.Use(middleware.Authorization(authManager))
	userSuperGroup.Use(middleware.RequireRole(dto.RoleSuperAdmin))
	{
		userSuperGroup.DELETE("/:id", userHandlers.DeleteUser)
		userSuperGroup.POST("/:id/restore", userHandlers.RestoreUser)
	}

	// Сборка зависимостей интересов
	interestRepo := repoInterests.NewInterestRepository(deps.DB.Pool, deps.DB.QueryTimeout)
	interestService := interestSvc.NewInterestService(deps.Logger, interestRepo)
	interestHandlers := interestH.NewInterestHandlers(deps.Logger, interestService)

	interestGroup := r.Group("/interests")
	interestGroup.Use(middleware.Authorization(authManager))
	{
		interestGroup.GET("", interestHandlers.ListInterests)
	}

	interestAdminGroup := r.Group("/interests")
	interestAdminGroup.Use(middleware.Authorization(authManager))
	interestAdminGroup.Use(middleware.RequireRole(dto.RoleSuperAdmin))
	{
		interestAdminGroup.POST("", interestHandlers.CreateInterest)
		interestAdminGroup.PATCH("/:id", interestHandlers.UpdateInterest)
		interestAdminGroup.DELETE("/:id", interestHandlers.DeleteInterest)
		interestAdminGroup.POST("/verify", interestHandlers.VerifyInterests)
	}

	return r
}
