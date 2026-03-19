package user

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type UserService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*dto.User, error)
	ListUsers(ctx context.Context, filter dto.UserFilter) ([]*dto.User, int, error)
	UpdateUser(ctx context.Context, user *dto.User) (*dto.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	SetActive(ctx context.Context, userID uuid.UUID, active bool) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	RestoreUser(ctx context.Context, id uuid.UUID) (*dto.User, error)
}
