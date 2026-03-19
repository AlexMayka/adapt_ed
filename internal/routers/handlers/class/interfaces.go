package class

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type ClassService interface {
	// GetClass возвращает данные класса по ID.
	GetClass(ctx context.Context, id uuid.UUID) (*dto.Class, error)

	// ListClasses возвращает список классов по фильтрам.
	ListClasses(ctx context.Context, filter dto.ClassFilter) ([]*dto.Class, int, error)

	// CreateClass создаёт новый класс.
	CreateClass(ctx context.Context, class *dto.Class) (*dto.Class, error)

	// UpdateClass обновляет данные класса.
	UpdateClass(ctx context.Context, class *dto.Class) (*dto.Class, error)

	// DeleteClass выполняет мягкое удаление класса.
	DeleteClass(ctx context.Context, id uuid.UUID) error

	// RestoreClass восстанавливает мягко удалённый класс.
	RestoreClass(ctx context.Context, id uuid.UUID) (*dto.Class, error)
}
