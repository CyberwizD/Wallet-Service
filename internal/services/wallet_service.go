package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CyberwizD/Wallet-Service/internal/models"
	"github.com/CyberwizD/Wallet-Service/internal/util"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalletService encapsulates wallet operations and Paystack integration.
type WalletService struct {
	db       *gorm.DB
	paystack *PaystackService
}

// LockClause serializes balance updates.
var LockClause = clause.Locking{Strength: "UPDATE"}

// NewWalletService constructs a WalletService.
func NewWalletService(db *gorm.DB, paystack *PaystackService) *WalletService {
	return &WalletService{db: db, paystack: paystack}
}

// InitiateDeposit records a pending transaction and returns a Paystack checkout URL.
func (s *WalletService) InitiateDeposit(user *models.User, amount int64) (string, string, error) {
	if amount <= 0 {
		return "", "", errors.New("amount must be greater than zero")
	}
	if user == nil || user.Wallet.ID == "" {
		return "", "", errors.New("wallet not found for user")
	}
	ref := fmt.Sprintf("DEP-%s", util.MustUUID())
	tx := models.Transaction{
		ID:        util.MustUUID(),
		Reference: ref,
		Type:      models.TransactionTypeDeposit,
		Status:    models.TransactionPending,
		Amount:    amount,
		WalletID:  user.Wallet.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.db.Create(&tx).Error; err != nil {
		return "", "", err
	}
	authURL, err := s.paystack.InitializeTransaction(amount, user.Email, ref)
	if err != nil {
		return "", "", err
	}
	return ref, authURL, nil
}

// ApplyDepositWebhook verifies idempotency and credits wallet on success.
func (s *WalletService) ApplyDepositWebhook(reference string, status string, payload []byte) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record models.Transaction
		if err := tx.Clauses(LockClause).First(&record, "reference = ?", reference).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("transaction reference not found")
			}
			return err
		}
		if record.Type != models.TransactionTypeDeposit {
			return errors.New("reference is not a deposit transaction")
		}
		if record.Status == models.TransactionSuccess {
			return nil // idempotent
		}
		var wallet models.Wallet
		if err := tx.Clauses(LockClause).First(&wallet, "id = ?", record.WalletID).Error; err != nil {
			return err
		}
		switch strings.ToLower(status) {
		case "success":
			wallet.Balance += record.Amount
			record.Status = models.TransactionSuccess
			record.RawPayload = payload
		case "failed":
			record.Status = models.TransactionFailed
			record.RawPayload = payload
		default:
			// do not update wallet for pending/unknown
			return nil
		}
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}
		if err := tx.Save(&record).Error; err != nil {
			return err
		}
		return nil
	})
}

// Transfer moves balance between two wallets atomically and records transactions.
func (s *WalletService) Transfer(sender *models.User, destWalletNumber string, amount int64) error {
	if sender == nil || sender.Wallet.ID == "" {
		return errors.New("sender wallet not found")
	}
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if sender.Wallet.Number == destWalletNumber {
		return errors.New("cannot transfer to the same wallet")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var senderWallet models.Wallet
		if err := tx.Clauses(LockClause).First(&senderWallet, "id = ?", sender.Wallet.ID).Error; err != nil {
			return err
		}
		if senderWallet.Balance < amount {
			return errors.New("insufficient balance")
		}
		var destWallet models.Wallet
		if err := tx.Clauses(LockClause).First(&destWallet, "number = ?", destWalletNumber).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("recipient wallet not found")
			}
			return err
		}
		senderWallet.Balance -= amount
		destWallet.Balance += amount

		now := time.Now()
		if err := tx.Save(&senderWallet).Error; err != nil {
			return err
		}
		if err := tx.Save(&destWallet).Error; err != nil {
			return err
		}
		senderTx := models.Transaction{
			ID:                 util.MustUUID(),
			Reference:          fmt.Sprintf("TRF-%s", util.MustUUID()),
			Type:               models.TransactionTypeTransfer,
			Status:             models.TransactionSuccess,
			Amount:             amount,
			WalletID:           senderWallet.ID,
			CounterpartyWallet: destWallet.Number,
			Description:        "debit transfer",
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		receiverTx := models.Transaction{
			ID:                 util.MustUUID(),
			Reference:          fmt.Sprintf("TRF-%s", util.MustUUID()),
			Type:               models.TransactionTypeTransfer,
			Status:             models.TransactionSuccess,
			Amount:             amount,
			WalletID:           destWallet.ID,
			CounterpartyWallet: senderWallet.Number,
			Description:        "credit transfer",
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := tx.Create(&senderTx).Error; err != nil {
			return err
		}
		if err := tx.Create(&receiverTx).Error; err != nil {
			return err
		}
		return nil
	})
}

// DepositStatus fetches a deposit transaction status.
func (s *WalletService) DepositStatus(reference string) (*models.Transaction, error) {
	var tx models.Transaction
	if err := s.db.First(&tx, "reference = ?", reference).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

// Balance returns the wallet balance for a user.
func (s *WalletService) Balance(userID string) (int64, error) {
	var wallet models.Wallet
	if err := s.db.First(&wallet, "user_id = ?", userID).Error; err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

// Transactions lists wallet transactions for a user ordered newest first.
func (s *WalletService) Transactions(userID string) ([]models.Transaction, error) {
	var wallet models.Wallet
	if err := s.db.First(&wallet, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	var list []models.Transaction
	if err := s.db.Where("wallet_id = ?", wallet.ID).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
