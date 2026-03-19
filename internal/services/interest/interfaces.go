package interest

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type InterestRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Interest, error)
	List(ctx context.Context, filter dto.InterestFilter) ([]*dto.Interest, int, error)
	Create(ctx context.Context, interest *dto.Interest) (*dto.Interest, error)
	Update(ctx context.Context, interest *dto.Interest) (*dto.Interest, error)
	Delete(ctx context.Context, id uuid.UUID) error
	VerifyBatch(ctx context.Context, ids []uuid.UUID) (int, error)
}
