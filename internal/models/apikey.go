package models

import "time"

// APIKey represents a service-to-service credential scoped to permissions.
type APIKey struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	UserID      string `gorm:"type:uuid;index"`
	User        User   `gorm:"constraint:OnDelete:CASCADE;"`
	Name        string
	KeyHash     string    `gorm:"uniqueIndex"`
	Permissions string    // comma separated list
	ExpiresAt   time.Time `gorm:"index"`
	Revoked     bool      `gorm:"index"`
	LastUsedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
