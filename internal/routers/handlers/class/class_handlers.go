package class

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	logInf "backend/internal/logger/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// ClassHandlers содержит HTTP-обработчики классов.
type ClassHandlers struct {
	log     logInf.Logger
	service ClassService
}

// NewClassHandlers создаёт обработчики классов.
func NewClassHandlers(log logInf.Logger, service ClassService) *ClassHandlers {
	return &ClassHandlers{log: log, service: service}
}

// toResponse конвертирует dto.Class в ClassResponse.
func toResponse(c *dto.Class) ClassResponse {
	return ClassResponse{
		ClassID:           ClassID{ID: c.ID},
		ClassSchoolID:     ClassSchoolID{SchoolID: c.SchoolID},
		ClassNumber:       ClassNumber{NumberOfClass: c.NumberOfClass},
		ClassSuffix:       ClassSuffix{SuffixesOfClass: c.SuffixesOfClass},
		ClassAcademicYear: ClassAcademicYear{AcademicYearStart: c.AcademicYearStart, AcademicYearEnd: c.AcademicYearEnd},
		ClassMeta:         ClassMeta{CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt},
	}
}

// parseSchoolID извлекает school_id из URL и проверяет доступ school_admin.
func parseSchoolID(c *gin.Context) (uuid.UUID, bool) {
	schoolID, err := uuid.Parse(c.Param("school_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID школы"))
		return uuid.Nil, false
	}

	roleVal, _ := c.Get(dto.CtxRole)
	callerRole, _ := roleVal.(dto.UserRole)

	if callerRole == dto.RoleSchoolAdmin {
		schoolVal, ok := c.Get(dto.CtxSchoolID)
		if !ok {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "школа администратора не определена"))
			return uuid.Nil, false
		}
		callerSchoolID, _ := schoolVal.(uuid.UUID)
		if callerSchoolID != schoolID {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "нет прав на управление классами этой школы"))
			return uuid.Nil, false
		}
	}

	return schoolID, true
}

// handleError обрабатывает AppError или generic error.
func handleError(c *gin.Context, err error) {
	if ae, ok := appErr.AsAppError(err); ok {
		c.JSON(ae.Status, dto.NewErrorResponse(c, ae.Code, ae.Message))
	} else {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, appErr.ErrCodeInternalServer, err.Error()))
	}
}

// GetClass     godoc
// @Summary      Получение класса
// @Description  Возвращает данные класса по ID.
// @Tags         classes
// @Produce      json
// @Security     BearerAuth
// @Param        school_id path string true "UUID школы"
// @Param        id        path string true "UUID класса"
// @Success      200 {object} ClassResponse
// @Failure      404 {object} dto.ErrorResponse "Класс не найден"
// @Router       /schools/{school_id}/classes/{id} [get]
func (h *ClassHandlers) GetClass(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID класса"))
		return
	}

	class, err := h.service.GetClass(c, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(class))
}

// ListClasses  godoc
// @Summary      Список классов школы
// @Description  Возвращает список классов школы с фильтрацией и пагинацией.
// @Tags         classes
// @Produce      json
// @Security     BearerAuth
// @Param        school_id      path  string false "UUID школы"
// @Param        number_of_class query int    false "Фильтр по номеру класса"
// @Param        limit          query int    false "Количество записей (по умолчанию 20, макс 100)"
// @Param        offset         query int    false "Смещение (по умолчанию 0)"
// @Success      200 {object} ListResponse
// @Router       /schools/{school_id}/classes [get]
func (h *ClassHandlers) ListClasses(c *gin.Context) {
	schoolID, err := uuid.Parse(c.Param("school_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID школы"))
		return
	}

	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	filter := dto.ClassFilter{
		SchoolID:      schoolID,
		NumberOfClass: req.NumberOfClass,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}

	classes, total, err := h.service.ListClasses(c, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	resp := ListResponse{
		Classes: make([]ClassResponse, 0, len(classes)),
		Total:   total,
	}
	for _, cl := range classes {
		resp.Classes = append(resp.Classes, toResponse(cl))
	}

	c.JSON(http.StatusOK, resp)
}

// CreateClass  godoc
// @Summary      Создание класса
// @Description  Создаёт новый класс в школе. Доступно super_admin и school_admin своей школы.
// @Tags         classes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        school_id path string        true "UUID школы"
// @Param        body      body CreateRequest true "Данные класса"
// @Success      201 {object} ClassResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      409 {object} dto.ErrorResponse "Класс с таким номером и суффиксом уже существует"
// @Router       /schools/{school_id}/classes [post]
func (h *ClassHandlers) CreateClass(c *gin.Context) {
	schoolID, ok := parseSchoolID(c)
	if !ok {
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	class := &dto.Class{
		SchoolID:          schoolID,
		NumberOfClass:     req.NumberOfClass,
		SuffixesOfClass:   req.SuffixesOfClass,
		AcademicYearStart: req.AcademicYearStart,
		AcademicYearEnd:   req.AcademicYearEnd,
	}

	created, err := h.service.CreateClass(c, class)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

// UpdateClass  godoc
// @Summary      Обновление класса
// @Description  Обновляет данные класса. Доступно super_admin и school_admin своей школы.
// @Tags         classes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        school_id path string        true "UUID школы"
// @Param        id        path string        true "UUID класса"
// @Param        body      body UpdateRequest true "Обновляемые поля"
// @Success      200 {object} ClassResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Класс не найден"
// @Failure      409 {object} dto.ErrorResponse "Класс с таким номером и суффиксом уже существует"
// @Router       /schools/{school_id}/classes/{id} [patch]
func (h *ClassHandlers) UpdateClass(c *gin.Context) {
	_, ok := parseSchoolID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID класса"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	class := &dto.Class{ID: id}
	if req.NumberOfClass != nil {
		class.NumberOfClass = *req.NumberOfClass
	}
	if req.SuffixesOfClass != nil {
		class.SuffixesOfClass = *req.SuffixesOfClass
	}
	class.AcademicYearStart = req.AcademicYearStart
	class.AcademicYearEnd = req.AcademicYearEnd

	updated, err := h.service.UpdateClass(c, class)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// DeleteClass  godoc
// @Summary      Удаление класса
// @Description  Мягкое удаление класса. Доступно super_admin и school_admin своей школы.
// @Tags         classes
// @Produce      json
// @Security     BearerAuth
// @Param        school_id path string true "UUID школы"
// @Param        id        path string true "UUID класса"
// @Success      200 {object} ClassID
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Класс не найден"
// @Router       /schools/{school_id}/classes/{id} [delete]
func (h *ClassHandlers) DeleteClass(c *gin.Context) {
	_, ok := parseSchoolID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID класса"))
		return
	}

	if err := h.service.DeleteClass(c, id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ClassID{ID: id})
}

// RestoreClass godoc
// @Summary      Восстановление класса
// @Description  Восстанавливает мягко удалённый класс. Доступно только super_admin.
// @Tags         classes
// @Produce      json
// @Security     BearerAuth
// @Param        school_id path string true "UUID школы"
// @Param        id        path string true "UUID класса"
// @Success      200 {object} ClassResponse
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Удалённый класс не найден"
// @Router       /schools/{school_id}/classes/{id}/restore [post]
func (h *ClassHandlers) RestoreClass(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID класса"))
		return
	}

	class, err := h.service.RestoreClass(c, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(class))
}
