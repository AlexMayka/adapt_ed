package types

import "time"

// UploadInfo содержит результат успешной загрузки объекта без привязки к типам MinIO SDK.
type UploadInfo struct {
	Bucket           string
	Key              string
	Tags             string
	Size             int64
	LastModified     time.Time
	Location         string
	VersionID        string
	Expiration       time.Time
	ExpirationRuleID string
}
