package user

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

// UserHandlers содержит HTTP-обработчики пользователей.
type UserHandlers struct {
	log     logInf.Logger
	service UserService
}

// NewUserHandlers создаёт обработчики пользователей.
func NewUserHandlers(log logInf.Logger, service UserService) *UserHandlers {
	return &UserHandlers{log: log, service: service}
}

func toResponse(u *dto.User) UserResponse {
	return UserResponse{
		UserID:        UserID{ID: u.ID},
		UserEmail:     UserEmail{Email: u.Email},
		UserFIO:       UserFIO{LastName: u.LastName, FirstName: u.FirstName, MiddleName: u.MiddleName},
		UserRole:      UserRole{Role: u.Role},
		UserEducation: UserEducation{ClassID: u.ClassID, SchoolID: u.SchoolID},
		UserAvatar:    UserAvatar{AvatarKey: u.AvatarKey},
		UserMeta:      UserMeta{IsActive: u.IsActive, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt},
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

// GetUser       godoc
// @Summary      Получение пользователя
// @Description  Возвращает данные пользователя по ID.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID пользователя"
// @Success      200 {object} UserResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandlers) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	user, err := h.service.GetUser(c, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(user))
}

// ListUsers     godoc
// @Summary      Список пользователей
// @Description  Возвращает список пользователей с фильтрацией и пагинацией. school_admin видит только свою школу.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        school_id query string false "Фильтр по школе"
// @Param        class_id  query string false "Фильтр по классу"
// @Param        role      query string false "Фильтр по роли"
// @Param        name      query string false "Поиск по ФИО"
// @Param        limit     query int    false "Количество записей"
// @Param        offset    query int    false "Смещение"
// @Success      200 {object} ListResponse
// @Router       /users [get]
func (h *UserHandlers) ListUsers(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	// school_admin принудительно фильтруется по своей школе
	roleVal, _ := c.Get(dto.CtxRole)
	callerRole, _ := roleVal.(dto.UserRole)

	filter := dto.UserFilter{
		SchoolID: req.SchoolID,
		ClassID:  req.ClassID,
		Role:     req.Role,
		Name:     req.Name,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	if callerRole == dto.RoleSchoolAdmin {
		schoolVal, ok := c.Get(dto.CtxSchoolID)
		if ok {
			sid, _ := schoolVal.(uuid.UUID)
			filter.SchoolID = &sid
		}
	}

	users, total, err := h.service.ListUsers(c, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	resp := ListResponse{
		Users: make([]UserResponse, 0, len(users)),
		Total: total,
	}
	for _, u := range users {
		resp.Users = append(resp.Users, toResponse(u))
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateProfile godoc
// @Summary      Обновление своего профиля
// @Description  Обновляет ФИО, email и аватар текущего пользователя.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body UpdateProfileRequest true "Обновляемые поля"
// @Success      200 {object} UserResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Email уже зарегистрирован"
// @Router       /users/me [patch]
func (h *UserHandlers) UpdateProfile(c *gin.Context) {
	userID, ok := getUserIDFromCtx(c)
	if !ok {
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	user := &dto.User{ID: userID}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	user.MiddleName = req.MiddleName
	user.AvatarKey = req.AvatarKey

	updated, err := h.service.UpdateUser(c, user)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// UpdateUser    godoc
// @Summary      Обновление пользователя админом
// @Description  Обновляет данные пользователя, включая привязку к школе и классу.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string          true "UUID пользователя"
// @Param        body body UpdateUserRequest true "Обновляемые поля"
// @Success      200 {object} UserResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Email уже зарегистрирован"
// @Router       /users/{id} [patch]
func (h *UserHandlers) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	// school_admin может обновлять только пользователей своей школы
	roleVal, _ := c.Get(dto.CtxRole)
	callerRole, _ := roleVal.(dto.UserRole)

	if callerRole == dto.RoleSchoolAdmin {
		target, err := h.service.GetUser(c, id)
		if err != nil {
			handleError(c, err)
			return
		}
		schoolVal, ok := c.Get(dto.CtxSchoolID)
		if !ok || target.SchoolID == nil {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "нет прав на обновление этого пользователя"))
			return
		}
		callerSchoolID, _ := schoolVal.(uuid.UUID)
		if callerSchoolID != *target.SchoolID {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(c, appErr.ErrCodeForbidden, "нет прав на обновление этого пользователя"))
			return
		}
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	user := &dto.User{ID: id}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	user.MiddleName = req.MiddleName
	user.AvatarKey = req.AvatarKey
	user.ClassID = req.ClassID
	user.SchoolID = req.SchoolID

	updated, err := h.service.UpdateUser(c, user)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// ChangePassword godoc
// @Summary      Смена пароля
// @Description  Меняет пароль текущего пользователя.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body ChangePasswordRequest true "Старый и новый пароль"
// @Success      200 {object} map[string]string "ok"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse "Неверный текущий пароль"
// @Router       /users/me/password [post]
func (h *UserHandlers) ChangePassword(c *gin.Context) {
	userID, ok := getUserIDFromCtx(c)
	if !ok {
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	if errs := utils.ValidatePassword(req.NewPassword); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, fmt.Sprintf("невалидный пароль: %s", errs)))
		return
	}

	if err := h.service.ChangePassword(c, userID, req.OldPassword, req.NewPassword); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SetActive     godoc
// @Summary      Активация/деактивация пользователя
// @Description  Устанавливает активность пользователя. Доступно super_admin и school_admin своей школы.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string           true "UUID пользователя"
// @Param        body body SetActiveRequest true "Активность"
// @Success      200 {object} map[string]string "ok"
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /users/{id}/active [patch]
func (h *UserHandlers) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	var req SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, err.Error()))
		return
	}

	if err := h.service.SetActive(c, id, req.IsActive); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteUser    godoc
// @Summary      Удаление пользователя
// @Description  Мягкое удаление. Доступно только super_admin.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID пользователя"
// @Success      200 {object} UserID
// @Failure      404 {object} dto.ErrorResponse
// @Router       /users/{id} [delete]
func (h *UserHandlers) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	if err := h.service.DeleteUser(c, id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, UserID{ID: id})
}

// RestoreUser   godoc
// @Summary      Восстановление пользователя
// @Description  Восстанавливает мягко удалённого пользователя. Доступно только super_admin.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID пользователя"
// @Success      200 {object} UserResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /users/{id}/restore [post]
func (h *UserHandlers) RestoreUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, appErr.ErrCodeBadRequest, "некорректный UUID"))
		return
	}

	user, err := h.service.RestoreUser(c, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(user))
}
