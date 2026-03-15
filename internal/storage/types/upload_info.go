package types

import "time"

// UploadInfo contains the result of a successful object upload without exposing MinIO SDK types.
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
