package dto

import (
	"github.com/google/uuid"
	"time"
)

// School содержит данные школы для передачи между слоями.
type School struct {
	ID        uuid.UUID
	Name      string
	City      string
	LogoKey   *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

// SchoolFilter параметры фильтрации и пагинации для списка школ.
type SchoolFilter struct {
	Name   *string
	City   *string
	Limit  int
	Offset int
}
