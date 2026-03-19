package profile

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.StudentProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, level *dto.DifficultyLevel, interests []uuid.UUID) (*dto.StudentProfile, error)
}
