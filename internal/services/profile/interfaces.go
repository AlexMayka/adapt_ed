package profile

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*dto.StudentProfile, error)
	Create(ctx context.Context, profile *dto.StudentProfile) (*dto.StudentProfile, error)
	DeactivateByUserID(ctx context.Context, userID uuid.UUID) error
}
