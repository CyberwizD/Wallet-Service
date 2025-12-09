package tests

import (
	"testing"

	"github.com/CyberwizD/Wallet-Service/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB spins up an in-memory sqlite DB and runs migrations.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.APIKey{}, &models.Transaction{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}
