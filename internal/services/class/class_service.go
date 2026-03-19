package class

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

// ClassService реализует бизнес-логику работы с классами.
type ClassService struct {
	log      interfaces.Logger
	classRep ClassRepository
}

// NewClassService создаёт сервис классов.
func NewClassService(log interfaces.Logger, classRep ClassRepository) *ClassService {
	return &ClassService{log: log, classRep: classRep}
}

// GetClass возвращает класс по ID.
func (s *ClassService) GetClass(ctx context.Context, id uuid.UUID) (*dto.Class, error) {
	class, err := s.classRep.GetByID(ctx, id)

	if errors.Is(err, appErr.ErrClassNotFound) {
		s.log.Info("класс не найден", "class_id", id)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "класс не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения класса", "err", err, "class_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return class, nil
}

// ListClasses возвращает список классов школы.
func (s *ClassService) ListClasses(ctx context.Context, filter dto.ClassFilter) ([]*dto.Class, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	classes, total, err := s.classRep.List(ctx, filter)
	if err != nil {
		s.log.Error("ошибка получения списка классов", "err", err, "school_id", filter.SchoolID)
		return nil, 0, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return classes, total, nil
}

// CreateClass создаёт новый класс.
func (s *ClassService) CreateClass(ctx context.Context, class *dto.Class) (*dto.Class, error) {
	now := time.Now()
	class.ID = utils.GetUniqUUID()
	class.CreatedAt = &now
	class.UpdatedAt = &now

	created, err := s.classRep.Create(ctx, class)

	if errors.Is(err, appErr.ErrClassAlreadyExists) {
		s.log.Info("класс уже существует", "school_id", class.SchoolID, "number", class.NumberOfClass, "suffix", class.SuffixesOfClass)
		return nil, appErr.NewAppError(http.StatusConflict, appErr.ErrClassAlreadyExists, "класс с таким номером и суффиксом уже существует")
	}
	if err != nil {
		s.log.Error("ошибка создания класса", "err", err)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка создания класса")
	}

	return created, nil
}

// UpdateClass обновляет данные класса.
func (s *ClassService) UpdateClass(ctx context.Context, class *dto.Class) (*dto.Class, error) {
	existing, err := s.classRep.GetByID(ctx, class.ID)

	if errors.Is(err, appErr.ErrClassNotFound) {
		s.log.Info("класс не найден для обновления", "class_id", class.ID)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "класс не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения класса", "err", err, "class_id", class.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if class.NumberOfClass != 0 {
		existing.NumberOfClass = class.NumberOfClass
	}
	if class.SuffixesOfClass != "" {
		existing.SuffixesOfClass = class.SuffixesOfClass
	}
	if class.AcademicYearStart != nil {
		existing.AcademicYearStart = class.AcademicYearStart
	}
	if class.AcademicYearEnd != nil {
		existing.AcademicYearEnd = class.AcademicYearEnd
	}

	updated, err := s.classRep.Update(ctx, existing)

	if errors.Is(err, appErr.ErrClassAlreadyExists) {
		return nil, appErr.NewAppError(http.StatusConflict, appErr.ErrClassAlreadyExists, "класс с таким номером и суффиксом уже существует")
	}
	if errors.Is(err, appErr.ErrClassNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "класс не найден")
	}
	if err != nil {
		s.log.Error("ошибка обновления класса", "err", err, "class_id", class.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления класса")
	}

	return updated, nil
}

// DeleteClass выполняет мягкое удаление класса.
func (s *ClassService) DeleteClass(ctx context.Context, id uuid.UUID) error {
	err := s.classRep.SoftDelete(ctx, id)

	if errors.Is(err, appErr.ErrClassNotFound) {
		s.log.Info("класс не найден для удаления", "class_id", id)
		return appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "класс не найден")
	}
	if err != nil {
		s.log.Error("ошибка удаления класса", "err", err, "class_id", id)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка удаления класса")
	}

	s.log.Info("класс удалён", "class_id", id)
	return nil
}

// RestoreClass восстанавливает мягко удалённый класс.
func (s *ClassService) RestoreClass(ctx context.Context, id uuid.UUID) (*dto.Class, error) {
	class, err := s.classRep.Restore(ctx, id)

	if errors.Is(err, appErr.ErrClassNotFound) {
		s.log.Info("класс не найден для восстановления", "class_id", id)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "удалённый класс не найден")
	}
	if err != nil {
		s.log.Error("ошибка восстановления класса", "err", err, "class_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка восстановления класса")
	}

	s.log.Info("класс восстановлен", "class_id", id)
	return class, nil
}
