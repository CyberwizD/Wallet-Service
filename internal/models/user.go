package models

import (
	"time"
)

// User represents an authenticated user in the system.
type User struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Wallet    Wallet
	APIKeys   []APIKey
}
