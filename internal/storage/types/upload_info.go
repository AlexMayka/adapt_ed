package types

import "time"

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
