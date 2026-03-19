package class

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type ClassRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Class, error)
	List(ctx context.Context, filter dto.ClassFilter) ([]*dto.Class, int, error)
	Create(ctx context.Context, class *dto.Class) (*dto.Class, error)
	Update(ctx context.Context, class *dto.Class) (*dto.Class, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) (*dto.Class, error)
}
