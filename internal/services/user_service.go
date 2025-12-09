package services

import (
	"errors"
	"time"

	"github.com/CyberwizD/Wallet-Service/internal/models"
	"github.com/CyberwizD/Wallet-Service/internal/util"

	"gorm.io/gorm"
)

// UserService owns user lifecycle operations.
type UserService struct {
	db *gorm.DB
}

// NewUserService constructs a UserService.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// UpsertGoogleUser ensures the user and wallet exist for a Google-authenticated account.
func (s *UserService) UpsertGoogleUser(email, name string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Wallet").First(&user, "email = ?", email).Error; err == nil {
		return &user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	walletNumber, err := s.generateWalletNumber()
	if err != nil {
		return nil, err
	}

	user = models.User{
		ID:        util.MustUUID(),
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	wallet := models.Wallet{
		ID:        util.MustUUID(),
		UserID:    user.ID,
		Number:    walletNumber,
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&wallet).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	user.Wallet = wallet
	return &user, nil
}

func (s *UserService) generateWalletNumber() (string, error) {
	for i := 0; i < 5; i++ {
		num, err := util.RandomDigits(12)
		if err != nil {
			return "", err
		}
		var count int64
		if err := s.db.Model(&models.Wallet{}).Where("number = ?", num).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return num, nil
		}
	}
	return "", errors.New("could not generate unique wallet number")
}
