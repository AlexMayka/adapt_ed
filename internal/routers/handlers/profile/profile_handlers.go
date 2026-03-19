package profile

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// ProfileHandlers содержит HTTP-обработчики профилей учеников.
type ProfileHandlers struct {
	log     logInf.Logger
	service ProfileService
}

// NewProfileHandlers создаёт обработчики профилей.
func NewProfileHandlers(log logInf.Logger, service ProfileService) *ProfileHandlers {
	return &ProfileHandlers{log: log, service: service}
}

func toResponse(p *dto.StudentProfile) ProfileResponse {
	return ProfileResponse{
		ID:           p.ID,
		UserID:       p.UserID,
		DefaultLevel: p.DefaultLevel,
		Interests:    p.Interests,
		Version:      p.Version,
		CreatedAt:    p.CreatedAt,
	}
}

func handleError(c *gin.Context, err error) {
	if ae, ok := appErr.AsAppError(err); ok {
		c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
	} else {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
	}
}

func getUserIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	val, ok := c.Get(dto.CtxUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(c, appErr.ErrCodeUnauthenticated, "идентификатор пользователя не найден в контексте"))
		return uuid.Nil, false
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, "некорректный тип идентификатора"))
		return uuid.Nil, false
	}
	return userID, true
}

// GetMyProfile godoc
// @Summary      Профиль ученика
// @Description  Возвращает текущий активный профиль ученика.
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ProfileResponse
// @Failure      404 {object} dto.ErrorResponse "Профиль не найден"
// @Router       /users/me/profile [get]
func (h *ProfileHandlers) GetMyProfile(c *gin.Context) {
	userID, ok := getUserIDFromCtx(c)
	if !ok {
		return
	}

	profile, err := h.service.GetProfile(c, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(profile))
}

// UpdateMyProfile godoc
// @Summary      Обновление профиля ученика
// @Description  Обновляет уровень сложности и/или интересы. Создаёт новую версию профиля.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body UpdateProfileRequest true "Обновляемые поля"
// @Success      200 {object} ProfileResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse "Профиль не найден"
// @Router       /users/me/profile [patch]
func (h *ProfileHandlers) UpdateMyProfile(c *gin.Context) {
	userID, ok := getUserIDFromCtx(c)
	if !ok {
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	profile, err := h.service.UpdateProfile(c, userID, req.DefaultLevel, req.Interests)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(profile))
}
