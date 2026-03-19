package interest

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// InterestHandlers содержит HTTP-обработчики интересов.
type InterestHandlers struct {
	log     logInf.Logger
	service InterestService
}

// NewInterestHandlers создаёт обработчики интересов.
func NewInterestHandlers(log logInf.Logger, service InterestService) *InterestHandlers {
	return &InterestHandlers{log: log, service: service}
}

func toResponse(i *dto.Interest) InterestResponse {
	return InterestResponse{
		InterestID: InterestID{ID: i.ID},
		Name:       i.Name,
		IconKey:    i.IconKey,
		IsVerified: i.IsVerified,
		CreatedAt:  i.CreatedAt,
	}
}

func handleError(c *gin.Context, err error) {
	if ae, ok := appErr.AsAppError(err); ok {
		c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
	} else {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
	}
}

// ListInterests godoc
// @Summary      Список интересов
// @Description  Возвращает список интересов с фильтрацией и пагинацией.
// @Tags         interests
// @Produce      json
// @Security     BearerAuth
// @Param        name        query string false "Поиск по названию"
// @Param        is_verified query bool   false "Фильтр по верификации"
// @Param        limit       query int    false "Количество записей"
// @Param        offset      query int    false "Смещение"
// @Success      200 {object} ListResponse
// @Router       /interests [get]
func (h *InterestHandlers) ListInterests(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	filter := dto.InterestFilter{
		Name:       req.Name,
		IsVerified: req.IsVerified,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	interests, total, err := h.service.ListInterests(c, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	resp := ListResponse{
		Interests: make([]InterestResponse, 0, len(interests)),
		Total:     total,
	}
	for _, i := range interests {
		resp.Interests = append(resp.Interests, toResponse(i))
	}

	c.JSON(http.StatusOK, resp)
}

// CreateInterest godoc
// @Summary      Создание интереса
// @Description  Создаёт новый верифицированный интерес. Доступно только super_admin.
// @Tags         interests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateRequest true "Данные интереса"
// @Success      201 {object} InterestResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Интерес уже существует"
// @Router       /interests [post]
func (h *InterestHandlers) CreateInterest(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	interest := &dto.Interest{Name: req.Name}

	created, err := h.service.CreateInterest(c, interest)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

// UpdateInterest godoc
// @Summary      Обновление интереса
// @Description  Обновляет название или иконку интереса. Доступно только super_admin.
// @Tags         interests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string        true "UUID интереса"
// @Param        body body UpdateRequest true "Обновляемые поля"
// @Success      200 {object} InterestResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Интерес с таким названием уже существует"
// @Router       /interests/{id} [patch]
func (h *InterestHandlers) UpdateInterest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	interest := &dto.Interest{ID: id}
	if req.Name != nil {
		interest.Name = *req.Name
	}
	interest.IconKey = req.IconKey

	updated, err := h.service.UpdateInterest(c, interest)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// DeleteInterest godoc
// @Summary      Удаление интереса
// @Description  Физически удаляет интерес из справочника. Доступно только super_admin.
// @Tags         interests
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID интереса"
// @Success      200 {object} InterestID
// @Failure      404 {object} dto.ErrorResponse
// @Router       /interests/{id} [delete]
func (h *InterestHandlers) DeleteInterest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	if err := h.service.DeleteInterest(c, id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, InterestID{ID: id})
}

// VerifyInterests godoc
// @Summary      Массовая верификация интересов
// @Description  Верифицирует интересы по списку ID. Доступно только super_admin.
// @Tags         interests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body VerifyRequest true "Список ID для верификации"
// @Success      200 {object} VerifyResponse
// @Failure      400 {object} dto.ErrorResponse
// @Router       /interests/verify [post]
func (h *InterestHandlers) VerifyInterests(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	count, err := h.service.VerifyInterests(c, req.IDs)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, VerifyResponse{Verified: count})
}
