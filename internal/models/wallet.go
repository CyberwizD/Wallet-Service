package models

import "time"

// Wallet stores the balance for a user in kobo (smallest currency unit).
type Wallet struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;uniqueIndex"`
	Number    string `gorm:"uniqueIndex;size:32"`
	Balance   int64  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
