package auth

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"backend/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// AuthHandlers содержит HTTP-обработчики авторизации.
type AuthHandlers struct {
	log     logInf.Logger
	service AuthService
}

// NewAuthHandlers создаёт обработчики авторизации.
func NewAuthHandlers(log logInf.Logger, service AuthService) *AuthHandlers {
	return &AuthHandlers{log: log, service: service}
}

// Registration   godoc
// @Summary      Самостоятельная регистрация
// @Description  Создаёт нового пользователя с ролью student и возвращает пару токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegistrationRequest true "Данные регистрации"
// @Success      201 {object} RegistrationResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Email уже зарегистрирован"
// @Router       /auth/registration [post]
func (h *AuthHandlers) Registration(c *gin.Context) {
	var req RegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	if errs := utils.ValidatePassword(req.Password); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, fmt.Sprintf("invalid password %s", errs)))
		return
	}

	userReg := dto.User{
		Email:      req.Email,
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
	}

	user, token, err := h.service.Registration(c, &userReg, req.Password, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, RegistrationResponse{
		UserID: UserID{ID: user.ID},
		UserBase: UserBase{
			AuthEmail: AuthEmail{Email: user.Email},
			Education: Education{
				ClassID:  user.ClassID,
				SchoolID: user.SchoolID,
			},
			FIO: FIO{
				LastName:   user.LastName,
				FirstName:  user.FirstName,
				MiddleName: user.MiddleName,
			},
		},
		Role: Role{Role: user.Role},
		UserMeta: UserMeta{
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AuthParamResponse: AuthParamResponse{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		},
	})
}

// RegistrationByAdmin godoc
// @Summary      Создание пользователя админом
// @Description  Создаёт пользователя с указанной ролью и генерирует временный пароль. Для school_admin school_id подставляется автоматически из токена.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body RegistrationRequestByAdmin true "Данные нового пользователя"
// @Success      201 {object} RegistrationResponseByAdmin
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      409 {object} dto.ErrorResponse "Email уже зарегистрирован"
// @Router       /auth/registration/admin [post]
func (h *AuthHandlers) RegistrationByAdmin(c *gin.Context) {
	var req RegistrationRequestByAdmin

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	roleVal, _ := c.Get(dto.CtxRole)
	callerRole, _ := roleVal.(dto.UserRole)

	var schoolID *uuid.UUID

	switch callerRole {
	case dto.RoleSchoolAdmin:
		if val, ok := c.Get(dto.CtxSchoolID); ok {
			if sid, ok := val.(uuid.UUID); ok {
				schoolID = &sid
			}
		}
		if schoolID == nil {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "у администратора школы не указана школа"))
			return
		}
	case dto.RoleSuperAdmin:
		schoolID = req.SchoolID
	}

	user := &dto.User{
		Email:      req.Email,
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		ClassID:    req.ClassID,
		SchoolID:   schoolID,
		Role:       req.Role.Role,
	}

	userDB, password, err := h.service.RegistrationByAdmin(c, user)
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, RegistrationResponseByAdmin{
		UserID: UserID{ID: userDB.ID},
		UserBase: UserBase{
			AuthEmail: AuthEmail{Email: userDB.Email},
			Education: Education{
				ClassID:  userDB.ClassID,
				SchoolID: userDB.SchoolID,
			},
			FIO: FIO{
				LastName:   userDB.LastName,
				FirstName:  userDB.FirstName,
				MiddleName: userDB.MiddleName,
			},
		},
		Role:              Role{Role: userDB.Role},
		GeneratedPassword: GeneratedPassword{Password: password},
		UserMeta: UserMeta{
			IsActive:  userDB.IsActive,
			CreatedAt: userDB.CreatedAt,
			UpdatedAt: userDB.UpdatedAt,
		},
	})
}

// Login          godoc
// @Summary      Аутентификация
// @Description  Выполняет вход по email и паролю, возвращает пару токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Учётные данные"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse "Неверный email или пароль"
// @Router       /auth/login [post]
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	user, tokens, err := h.service.Login(c.Request.Context(), req.Email, req.Password, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		UserID: UserID{ID: user.ID},
		UserBase: UserBase{
			AuthEmail: AuthEmail{Email: user.Email},
			Education: Education{
				ClassID:  user.ClassID,
				SchoolID: user.SchoolID,
			},
			FIO: FIO{
				LastName:   user.LastName,
				FirstName:  user.FirstName,
				MiddleName: user.MiddleName,
			},
		},
		UserMeta: UserMeta{
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Role: Role{Role: user.Role},
		AuthParamResponse: AuthParamResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	})
}

// GetMe          godoc
// @Summary      Текущий пользователь
// @Description  Возвращает данные авторизованного пользователя по JWT.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} GetMeResponse
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Router       /auth/me [get]
func (h *AuthHandlers) GetMe(c *gin.Context) {
	val, ok := c.Get(dto.CtxUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(c, appErr.ErrJWTInvalid, "идентификатор пользователя не найден в контексте"))
		return
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, "некорректный тип идентификатора пользователя"))
		return
	}

	user, err := h.service.GetMe(c, userID)

	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, GetMeResponse{
		UserID: UserID{ID: user.ID},
		UserBase: UserBase{
			AuthEmail: AuthEmail{Email: user.Email},
			Education: Education{
				ClassID:  user.ClassID,
				SchoolID: user.SchoolID,
			},
			FIO: FIO{
				LastName:   user.LastName,
				FirstName:  user.FirstName,
				MiddleName: user.MiddleName,
			},
		},
		UserMeta: UserMeta{
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Role:   Role{Role: user.Role},
		Avatar: Avatar{AvatarKey: user.AvatarKey},
	})
}

// Refresh        godoc
// @Summary      Обновление токенов
// @Description  Принимает refresh token и user_id, возвращает новую пару access + refresh.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh token и ID пользователя"
// @Success      200 {object} RefreshResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse "Невалидный или истёкший refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandlers) Refresh(c *gin.Context) {
	var req RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	tokens, err := h.service.Refresh(c, req.UserID, req.RefreshToken, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{
		AuthParamResponse: AuthParamResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	})
}

// Logout         godoc
// @Summary      Выход из текущей сессии
// @Description  Инвалидирует переданный refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body LogoutRequest true "Refresh token текущей сессии"
// @Success      200 {object} LogoutResponse
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Router       /auth/logout [post]
func (h *AuthHandlers) Logout(c *gin.Context) {
	val, ok := c.Get(dto.CtxUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(c, appErr.ErrCodeUnauthenticated, "идентификатор пользователя не найден в контексте"))
		return
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, "некорректный тип идентификатора пользователя"))
		return
	}

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	if err := h.service.Logout(c, userID, req.RefreshToken); err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, LogoutResponse{
		UserID: UserID{ID: userID},
	})
}

// LogoutAll      godoc
// @Summary      Выход со всех устройств
// @Description  Инвалидирует все refresh token текущего пользователя и увеличивает версию сессии.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} LogoutResponse
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Router       /auth/logout-all [post]
func (h *AuthHandlers) LogoutAll(c *gin.Context) {
	val, ok := c.Get(dto.CtxUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(c, appErr.ErrCodeUnauthenticated, "идентификатор пользователя не найден в контексте"))
		return
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, "некорректный тип идентификатора пользователя"))
		return
	}

	if err := h.service.LogoutAll(c, userID); err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, LogoutResponse{
		UserID: UserID{ID: userID},
	})
}
