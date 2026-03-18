package auth

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"backend/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
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
// @Description  Создаёт пользователя с указанной ролью и генерирует временный пароль.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegistrationRequestByAdmin true "Данные нового пользователя"
// @Success      201 {object} RegistrationResponseByAdmin
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      409 {object} dto.ErrorResponse "Email уже зарегистрирован"
// @Router       /auth/registration/admin [post]
func (h *AuthHandlers) RegistrationByAdmin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.NewErrorResponse(c, appErr.ErrCodeNotImplemented, "not implemented"))
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
	c.JSON(http.StatusNotImplemented, dto.NewErrorResponse(c, appErr.ErrCodeNotImplemented, "not implemented"))
}

// Refresh        godoc
// @Summary      Обновление токенов
// @Description  Принимает refresh token и возвращает новую пару access + refresh.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh token"
// @Success      200 {object} RefreshResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse "Невалидный или истёкший refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandlers) Refresh(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.NewErrorResponse(c, appErr.ErrCodeNotImplemented, "not implemented"))
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
	c.JSON(http.StatusNotImplemented, dto.NewErrorResponse(c, appErr.ErrCodeNotImplemented, "not implemented"))
}

// LogoutAll      godoc
// @Summary      Выход со всех устройств
// @Description  Инвалидирует все refresh token текущего пользователя.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} LogoutResponse
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Router       /auth/logout-all [post]
func (h *AuthHandlers) LogoutAll(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.NewErrorResponse(c, appErr.ErrCodeNotImplemented, "not implemented"))
}
