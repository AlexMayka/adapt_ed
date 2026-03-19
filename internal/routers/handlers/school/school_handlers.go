package school

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// SchoolHandlers содержит HTTP-обработчики школ.
type SchoolHandlers struct {
	log     logInf.Logger
	service SchoolService
}

// NewSchoolHandlers создаёт обработчики школ.
func NewSchoolHandlers(log logInf.Logger, service SchoolService) *SchoolHandlers {
	return &SchoolHandlers{log: log, service: service}
}

// toResponse конвертирует dto.School в SchoolResponse.
func toResponse(s *dto.School) SchoolResponse {
	return SchoolResponse{
		SchoolID:   SchoolID{ID: s.ID},
		SchoolBase: SchoolBase{SchoolName: SchoolName{Name: s.Name}, SchoolCity: SchoolCity{City: s.City}},
		Logo:       Logo{LogoKey: s.LogoKey},
		SchoolMeta: SchoolMeta{CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt},
	}
}

// GetSchool   godoc
// @Summary      Получение школы
// @Description  Возвращает данные школы по ID.
// @Tags         schools
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID школы"
// @Success      200 {object} SchoolResponse
// @Failure      404 {object} dto.ErrorResponse "Школа не найдена"
// @Router       /schools/{id} [get]
func (h *SchoolHandlers) GetSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	school, err := h.service.GetSchool(c, id)
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, toResponse(school))
}

// ListSchools  godoc
// @Summary      Список школ
// @Description  Возвращает список школ с фильтрацией и пагинацией.
// @Tags         schools
// @Produce      json
// @Security     BearerAuth
// @Param        name   query string false "Поиск по названию (подстрока)"
// @Param        city   query string false "Фильтр по городу"
// @Param        limit  query int    false "Количество записей (по умолчанию 20, макс 100)"
// @Param        offset query int    false "Смещение (по умолчанию 0)"
// @Success      200 {object} ListResponse
// @Router       /schools [get]
func (h *SchoolHandlers) ListSchools(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	filter := dto.SchoolFilter{
		Name:   req.Name,
		City:   req.City,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	schools, total, err := h.service.ListSchools(c, filter)
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	resp := ListResponse{
		Schools: make([]SchoolResponse, 0, len(schools)),
		Total:   total,
	}
	for _, s := range schools {
		resp.Schools = append(resp.Schools, toResponse(s))
	}

	c.JSON(http.StatusOK, resp)
}

// CreateSchool godoc
// @Summary      Создание школы
// @Description  Создаёт новую школу. Доступно только super_admin.
// @Tags         schools
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateRequest true "Данные школы"
// @Success      201 {object} SchoolResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Router       /schools [post]
func (h *SchoolHandlers) CreateSchool(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	school := &dto.School{
		Name: req.Name,
		City: req.City,
	}

	created, err := h.service.CreateSchool(c, school)
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

// UpdateSchool godoc
// @Summary      Обновление школы
// @Description  Обновляет данные школы. Доступно super_admin и school_admin своей школы.
// @Tags         schools
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string        true "UUID школы"
// @Param        body body UpdateRequest true "Обновляемые поля"
// @Success      200 {object} SchoolResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Школа не найдена"
// @Router       /schools/{id} [patch]
func (h *SchoolHandlers) UpdateSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	// school_admin может обновлять только свою школу
	roleVal, _ := c.Get(dto.CtxRole)
	callerRole, _ := roleVal.(dto.UserRole)

	if callerRole == dto.RoleSchoolAdmin {
		schoolVal, ok := c.Get(dto.CtxSchoolID)
		if !ok {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "школа администратора не определена"))
			return
		}
		callerSchoolID, _ := schoolVal.(uuid.UUID)
		if callerSchoolID != id {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "нет прав на обновление этой школы"))
			return
		}
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	school := &dto.School{ID: id}
	if req.Name != nil {
		school.Name = *req.Name
	}
	if req.City != nil {
		school.City = *req.City
	}
	school.LogoKey = req.LogoKey

	updated, err := h.service.UpdateSchool(c, school)
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// DeleteSchool godoc
// @Summary      Удаление школы
// @Description  Мягкое удаление школы. Доступно только super_admin.
// @Tags         schools
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID школы"
// @Success      200 {object} SchoolID
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Школа не найдена"
// @Router       /schools/{id} [delete]
func (h *SchoolHandlers) DeleteSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	if err := h.service.DeleteSchool(c, id); err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, SchoolID{ID: id})
}

// RestoreSchool godoc
// @Summary      Восстановление школы
// @Description  Восстанавливает мягко удалённую школу. Доступно только super_admin.
// @Tags         schools
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID школы"
// @Success      200 {object} SchoolResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Удалённая школа не найдена"
// @Router       /schools/{id}/restore [post]
func (h *SchoolHandlers) RestoreSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	school, err := h.service.RestoreSchool(c, id)
	if err != nil {
		if ae, ok := appErr.AsAppError(err); ok {
			c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, toResponse(school))
}
