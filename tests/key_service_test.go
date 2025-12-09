package tests

import (
	"testing"

	"github.com/CyberwizD/Wallet-Service/internal/models"
	"github.com/CyberwizD/Wallet-Service/internal/services"
	"github.com/CyberwizD/Wallet-Service/internal/util"
)

func TestCreateKeyEnforcesMaxActive(t *testing.T) {
	db := newTestDB(t)
	user := models.User{ID: util.MustUUID(), Email: "max@test.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	service := services.NewAPIKeyService(db)
	for i := 0; i < 5; i++ {
		if _, _, err := service.CreateKey(&user, "svc", []string{"deposit", "read"}, "1D"); err != nil {
			t.Fatalf("create key %d failed: %v", i, err)
		}
	}
	if _, _, err := service.CreateKey(&user, "svc", []string{"deposit"}, "1D"); err == nil {
		t.Fatalf("expected error when creating 6th key, got nil")
	}
}
