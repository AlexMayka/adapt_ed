package school

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type SchoolRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.School, error)
	List(ctx context.Context, filter dto.SchoolFilter) ([]*dto.School, int, error)
	Create(ctx context.Context, school *dto.School) (*dto.School, error)
	Update(ctx context.Context, school *dto.School) (*dto.School, error)
	Restore(ctx context.Context, id uuid.UUID) (*dto.School, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
