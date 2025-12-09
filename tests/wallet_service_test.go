package tests

import (
	"testing"

	"github.com/CyberwizD/Wallet-Service/internal/models"
	"github.com/CyberwizD/Wallet-Service/internal/services"
	"github.com/CyberwizD/Wallet-Service/internal/util"

	"gorm.io/gorm"
)

func seedUserWithWallet(db *gorm.DB, email string, balance int64) models.User {
	user := models.User{ID: util.MustUUID(), Email: email}
	wallet := models.Wallet{
		ID:      util.MustUUID(),
		UserID:  user.ID,
		Number:  util.MustUUID(),
		Balance: balance,
	}
	_ = db.Create(&user).Error
	_ = db.Create(&wallet).Error
	user.Wallet = wallet
	return user
}

func TestTransferSuccess(t *testing.T) {
	db := newTestDB(t)
	svc := services.NewWalletService(db, nil)

	sender := seedUserWithWallet(db, "sender@test.com", 10_000)
	receiver := seedUserWithWallet(db, "receiver@test.com", 2_000)

	if err := svc.Transfer(&sender, receiver.Wallet.Number, 5_000); err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	var updatedSender models.Wallet
	var updatedReceiver models.Wallet
	_ = db.First(&updatedSender, "id = ?", sender.Wallet.ID).Error
	_ = db.First(&updatedReceiver, "id = ?", receiver.Wallet.ID).Error
	if updatedSender.Balance != 5_000 {
		t.Fatalf("expected sender balance 5000, got %d", updatedSender.Balance)
	}
	if updatedReceiver.Balance != 7_000 {
		t.Fatalf("expected receiver balance 7000, got %d", updatedReceiver.Balance)
	}
}

func TestTransferInsufficientBalance(t *testing.T) {
	db := newTestDB(t)
	svc := services.NewWalletService(db, nil)

	sender := seedUserWithWallet(db, "sender2@test.com", 1_000)
	receiver := seedUserWithWallet(db, "receiver2@test.com", 0)

	if err := svc.Transfer(&sender, receiver.Wallet.Number, 5_000); err == nil {
		t.Fatalf("expected insufficient balance error")
	}
}
