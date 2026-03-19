package school

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type SchoolService interface {
	// GetSchool возвращает данные школы по ID.
	GetSchool(ctx context.Context, id uuid.UUID) (*dto.School, error)

	// ListSchools возвращает список школ по фильтрам.
	ListSchools(ctx context.Context, filter dto.SchoolFilter) ([]*dto.School, int, error)

	// CreateSchool создаёт новую школу.
	CreateSchool(ctx context.Context, school *dto.School) (*dto.School, error)

	// UpdateSchool обновляет данные школы и возвращает обновлённую запись.
	UpdateSchool(ctx context.Context, school *dto.School) (*dto.School, error)

	// RestoreSchool восстанавливает мягко удалённую школу.
	RestoreSchool(ctx context.Context, id uuid.UUID) (*dto.School, error)

	// DeleteSchool выполняет мягкое удаление школы.
	DeleteSchool(ctx context.Context, id uuid.UUID) error
}
