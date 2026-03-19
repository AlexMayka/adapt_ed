package interest

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type InterestService interface {
	GetInterest(ctx context.Context, id uuid.UUID) (*dto.Interest, error)
	ListInterests(ctx context.Context, filter dto.InterestFilter) ([]*dto.Interest, int, error)
	CreateInterest(ctx context.Context, interest *dto.Interest) (*dto.Interest, error)
	UpdateInterest(ctx context.Context, interest *dto.Interest) (*dto.Interest, error)
	DeleteInterest(ctx context.Context, id uuid.UUID) error
	VerifyInterests(ctx context.Context, ids []uuid.UUID) (int, error)
}
