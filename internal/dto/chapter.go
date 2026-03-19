package dto

import (
	"github.com/google/uuid"
	"time"
)

// Chapter содержит данные главы учебной программы (версионируется).
type Chapter struct {
	ID        uuid.UUID
	ProgramID uuid.UUID
	Title     string
	SortOrder int
	IconKey   *string
	IsActive  bool
	Version   int
	CreatedAt *time.Time
}

// ChapterFilter параметры фильтрации для списка глав.
type ChapterFilter struct {
	ProgramID uuid.UUID
	Limit   int
	Offset  int
}
