package models

import (
	"time"
)

// TransactionType enumerates supported transaction categories.
type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus captures lifecycle states for transactions.
type TransactionStatus string

const (
	TransactionPending TransactionStatus = "pending"
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
)

// Transaction represents any balance-impacting operation.
type Transaction struct {
	ID                 string            `gorm:"type:uuid;primaryKey"`
	Reference          string            `gorm:"uniqueIndex"`
	Type               TransactionType   `gorm:"index"`
	Status             TransactionStatus `gorm:"index"`
	Amount             int64
	WalletID           string `gorm:"index"`
	CounterpartyWallet string // recipient for transfers
	Description        string
	RawPayload         []byte `gorm:"type:jsonb"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
