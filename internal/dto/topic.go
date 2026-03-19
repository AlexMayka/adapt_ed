package dto

import (
	"github.com/google/uuid"
	"time"
)

// Topic содержит данные параграфа (версионируется).
type Topic struct {
	ID        uuid.UUID
	ChapterID uuid.UUID
	Title     string
	SortOrder int
	IconKey   *string
	IsActive  bool
	Version   int
	CreatedAt *time.Time
}

// TopicFilter параметры фильтрации для списка параграфов.
type TopicFilter struct {
	ChapterID uuid.UUID
	Limit     int
	Offset    int
}
