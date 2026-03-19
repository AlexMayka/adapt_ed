package school

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

// SchoolService реализует бизнес-логику работы со школами.
type SchoolService struct {
	log       interfaces.Logger
	schoolRep SchoolRepository
}

// NewSchoolService создаёт сервис школ.
func NewSchoolService(log interfaces.Logger, schoolRep SchoolRepository) *SchoolService {
	return &SchoolService{log: log, schoolRep: schoolRep}
}

// GetSchool возвращает школу по ID.
func (s *SchoolService) GetSchool(ctx context.Context, id uuid.UUID) (*dto.School, error) {
	school, err := s.schoolRep.GetByID(ctx, id)

	if errors.Is(err, appErr.ErrSchoolNotFound) {
		s.log.Info("школа не найдена", "school_id", id)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "школа не найдена")
	}

	if err != nil {
		s.log.Error("ошибка получения школы", "err", err, "school_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return school, nil
}

// ListSchools возвращает список школ по фильтрам.
func (s *SchoolService) ListSchools(ctx context.Context, filter dto.SchoolFilter) ([]*dto.School, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	schools, total, err := s.schoolRep.List(ctx, filter)
	if err != nil {
		s.log.Error("ошибка получения списка школ", "err", err)
		return nil, 0, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return schools, total, nil
}

// CreateSchool создаёт новую школу.
func (s *SchoolService) CreateSchool(ctx context.Context, school *dto.School) (*dto.School, error) {
	now := time.Now()
	school.ID = utils.GetUniqUUID()
	school.CreatedAt = &now
	school.UpdatedAt = &now

	created, err := s.schoolRep.Create(ctx, school)
	if err != nil {
		s.log.Error("ошибка создания школы", "err", err)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка создания школы")
	}

	return created, nil
}

// UpdateSchool обновляет данные школы.
func (s *SchoolService) UpdateSchool(ctx context.Context, school *dto.School) (*dto.School, error) {
	existing, err := s.schoolRep.GetByID(ctx, school.ID)

	if errors.Is(err, appErr.ErrSchoolNotFound) {
		s.log.Info("школа не найдена для обновления", "school_id", school.ID)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "школа не найдена")
	}

	if err != nil {
		s.log.Error("ошибка получения школы", "err", err, "school_id", school.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if school.Name != "" {
		existing.Name = school.Name
	}
	if school.City != "" {
		existing.City = school.City
	}
	if school.LogoKey != nil {
		existing.LogoKey = school.LogoKey
	}

	updated, err := s.schoolRep.Update(ctx, existing)

	if errors.Is(err, appErr.ErrSchoolNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "школа не найдена")
	}

	if err != nil {
		s.log.Error("ошибка обновления школы", "err", err, "school_id", school.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления школы")
	}

	return updated, nil
}

// DeleteSchool выполняет мягкое удаление школы.
func (s *SchoolService) DeleteSchool(ctx context.Context, id uuid.UUID) error {
	err := s.schoolRep.SoftDelete(ctx, id)

	if errors.Is(err, appErr.ErrSchoolNotFound) {
		s.log.Info("школа не найдена для удаления", "school_id", id)
		return appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "школа не найдена")
	}

	if err != nil {
		s.log.Error("ошибка удаления школы", "err", err, "school_id", id)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка удаления школы")
	}

	s.log.Info("школа удалена", "school_id", id)
	return nil
}
