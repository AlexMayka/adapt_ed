package profile

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"context"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// ProfileService реализует бизнес-логику профилей учеников.
type ProfileService struct {
	log        interfaces.Logger
	profileRep ProfileRepository
}

// NewProfileService создаёт сервис профилей.
func NewProfileService(log interfaces.Logger, profileRep ProfileRepository) *ProfileService {
	return &ProfileService{log: log, profileRep: profileRep}
}

// GetProfile возвращает активный профиль ученика.
func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.StudentProfile, error) {
	profile, err := s.profileRep.GetActiveByUserID(ctx, userID)

	if errors.Is(err, appErr.ErrProfileNotFound) {
		s.log.Info("профиль ученика не найден", "user_id", userID)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "профиль ученика не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения профиля", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return profile, nil
}

// CreateDefault создаёт профиль по умолчанию для нового студента.
func (s *ProfileService) CreateDefault(ctx context.Context, userID uuid.UUID) (*dto.StudentProfile, error) {
	now := time.Now()
	profile := &dto.StudentProfile{
		ID:           utils.GetUniqUUID(),
		UserID:       userID,
		DefaultLevel: dto.LevelSimple,
		Interests:    []uuid.UUID{},
		IsActive:     true,
		CreatedAt:    &now,
	}

	created, err := s.profileRep.Create(ctx, profile)
	if err != nil {
		s.log.Error("ошибка создания профиля ученика", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка создания профиля")
	}

	return created, nil
}

// UpdateProfile обновляет профиль ученика (создаёт новую версию).
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, level *dto.DifficultyLevel, interests []uuid.UUID) (*dto.StudentProfile, error) {
	existing, err := s.profileRep.GetActiveByUserID(ctx, userID)

	if errors.Is(err, appErr.ErrProfileNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "профиль ученика не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения профиля", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	newLevel := existing.DefaultLevel
	if level != nil {
		newLevel = *level
	}

	newInterests := existing.Interests
	if interests != nil {
		newInterests = interests
	}

	// Деактивируем текущую версию
	if err := s.profileRep.DeactivateByUserID(ctx, userID); err != nil {
		s.log.Error("ошибка деактивации профиля", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления профиля")
	}

	// Создаём новую версию
	now := time.Now()
	newProfile := &dto.StudentProfile{
		ID:           utils.GetUniqUUID(),
		UserID:       userID,
		DefaultLevel: newLevel,
		Interests:    newInterests,
		IsActive:     true,
		CreatedAt:    &now,
	}

	created, err := s.profileRep.Create(ctx, newProfile)
	if err != nil {
		s.log.Error("ошибка создания новой версии профиля", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления профиля")
	}

	s.log.Info("профиль ученика обновлён", "user_id", userID, "version", created.Version)
	return created, nil
}
