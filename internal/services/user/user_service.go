package user

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"context"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

// UserService реализует бизнес-логику работы с пользователями.
type UserService struct {
	log     interfaces.Logger
	userRep UserRepository
}

// NewUserService создаёт сервис пользователей.
func NewUserService(log interfaces.Logger, userRep UserRepository) *UserService {
	return &UserService{log: log, userRep: userRep}
}

// GetUser возвращает пользователя по ID.
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	user, err := s.userRep.GetUserByID(ctx, id)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден", "user_id", id)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения пользователя", "err", err, "user_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return user, nil
}

// ListUsers возвращает список пользователей по фильтрам.
func (s *UserService) ListUsers(ctx context.Context, filter dto.UserFilter) ([]*dto.User, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	users, total, err := s.userRep.List(ctx, filter)
	if err != nil {
		s.log.Error("ошибка получения списка пользователей", "err", err)
		return nil, 0, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return users, total, nil
}

// UpdateUser обновляет профиль пользователя.
func (s *UserService) UpdateUser(ctx context.Context, user *dto.User) (*dto.User, error) {
	existing, err := s.userRep.GetUserByID(ctx, user.ID)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден для обновления", "user_id", user.ID)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения пользователя", "err", err, "user_id", user.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if user.Email != "" {
		existing.Email = user.Email
	}
	if user.LastName != "" {
		existing.LastName = user.LastName
	}
	if user.FirstName != "" {
		existing.FirstName = user.FirstName
	}
	if user.MiddleName != nil {
		existing.MiddleName = user.MiddleName
	}
	if user.AvatarKey != nil {
		existing.AvatarKey = user.AvatarKey
	}
	if user.ClassID != nil {
		existing.ClassID = user.ClassID
	}
	if user.SchoolID != nil {
		existing.SchoolID = user.SchoolID
	}

	updated, err := s.userRep.Update(ctx, existing)

	if errors.Is(err, appErr.ErrEmailAlreadyExists) {
		return nil, appErr.NewAppError(http.StatusConflict, appErr.ErrEmailAlreadyExists, "email уже зарегистрирован")
	}
	if errors.Is(err, appErr.ErrUserNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка обновления пользователя", "err", err, "user_id", user.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления пользователя")
	}

	return updated, nil
}

// ChangePassword меняет пароль пользователя.
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRep.GetUserByID(ctx, userID)
	if errors.Is(err, appErr.ErrUserNotFound) {
		return appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения пользователя", "err", err, "user_id", userID)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if user.PasswordHash == nil || !utils.CheckValuesHash(oldPassword, *user.PasswordHash) {
		s.log.Info("неверный старый пароль", "user_id", userID)
		return appErr.NewAppError(http.StatusUnauthorized, appErr.ErrInvalidCredentials, "неверный текущий пароль")
	}

	newHash, err := utils.HashValue(newPassword)
	if err != nil {
		s.log.Error("ошибка хэширования пароля", "err", err)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка хэширования пароля")
	}

	if err := s.userRep.UpdatePassword(ctx, userID, newHash); err != nil {
		s.log.Error("ошибка обновления пароля", "err", err, "user_id", userID)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления пароля")
	}

	s.log.Info("пароль обновлён", "user_id", userID)
	return nil
}

// SetActive устанавливает активность пользователя.
func (s *UserService) SetActive(ctx context.Context, userID uuid.UUID, active bool) error {
	err := s.userRep.SetActive(ctx, userID, active)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден для изменения активности", "user_id", userID)
		return appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка изменения активности", "err", err, "user_id", userID)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка изменения активности")
	}

	return nil
}

// DeleteUser выполняет мягкое удаление пользователя.
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := s.userRep.SoftDelete(ctx, id)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден для удаления", "user_id", id)
		return appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка удаления пользователя", "err", err, "user_id", id)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка удаления пользователя")
	}

	s.log.Info("пользователь удалён", "user_id", id)
	return nil
}

// RestoreUser восстанавливает мягко удалённого пользователя.
func (s *UserService) RestoreUser(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	user, err := s.userRep.Restore(ctx, id)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден для восстановления", "user_id", id)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "удалённый пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка восстановления пользователя", "err", err, "user_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка восстановления пользователя")
	}

	s.log.Info("пользователь восстановлен", "user_id", id)
	return user, nil
}
