package dto

import (
	"github.com/google/uuid"
	"time"
)

// Subject содержит данные предмета.
type Subject struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	IconKey   *string
	Color     *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// SubjectFilter параметры фильтрации для списка предметов.
type SubjectFilter struct {
	Name   *string
	Limit  int
	Offset int
}
