package user

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*dto.User, error)
	List(ctx context.Context, filter dto.UserFilter) ([]*dto.User, int, error)
	Update(ctx context.Context, user *dto.User) (*dto.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	SetActive(ctx context.Context, userID uuid.UUID, active bool) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) (*dto.User, error)
}
