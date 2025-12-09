package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/CyberwizD/Wallet-Service/internal/config"
	"github.com/CyberwizD/Wallet-Service/internal/models"
	"github.com/CyberwizD/Wallet-Service/internal/util"

	"gorm.io/gorm"
)

// APIKeyService manages service-to-service credentials.
type APIKeyService struct {
	db *gorm.DB
}

// NewAPIKeyService constructs an APIKeyService.
func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

// CreateKey issues a new API key enforcing permission and count constraints.
func (s *APIKeyService) CreateKey(user *models.User, name string, permissions []string, expiry string) (*models.APIKey, string, error) {
	if user == nil {
		return nil, "", errors.New("user required")
	}
	if err := validatePermissions(permissions); err != nil {
		return nil, "", err
	}
	var activeCount int64
	if err := s.db.Model(&models.APIKey{}).
		Where("user_id = ? AND revoked = false AND expires_at > ?", user.ID, time.Now()).
		Count(&activeCount).Error; err != nil {
		return nil, "", err
	}
	if activeCount >= 5 {
		return nil, "", errors.New("maximum of 5 active API keys reached")
	}
	duration, err := config.ParseExpiry(expiry)
	if err != nil {
		return nil, "", errors.New("invalid expiry format; use 1H,1D,1M,1Y")
	}
	plainKey, err := util.GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}
	hash := sha256.Sum256([]byte(plainKey))
	record := models.APIKey{
		ID:          util.MustUUID(),
		UserID:      user.ID,
		Name:        name,
		KeyHash:     hex.EncodeToString(hash[:]),
		Permissions: util.PermissionsString(permissions),
		ExpiresAt:   time.Now().Add(duration),
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, "", err
	}
	return &record, plainKey, nil
}

// RolloverKey clones an expired key's permissions into a new key with a fresh expiry.
func (s *APIKeyService) RolloverKey(userID, expiredKeyID, expiry string) (*models.APIKey, string, error) {
	var expired models.APIKey
	if err := s.db.First(&expired, "id = ? AND user_id = ?", expiredKeyID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("expired key not found")
		}
		return nil, "", err
	}
	if time.Now().Before(expired.ExpiresAt) && !expired.Revoked {
		return nil, "", errors.New("key has not expired yet")
	}
	if err := validatePermissions(util.SplitPermissions(expired.Permissions)); err != nil {
		return nil, "", err
	}
	duration, err := config.ParseExpiry(expiry)
	if err != nil {
		return nil, "", errors.New("invalid expiry format; use 1H,1D,1M,1Y")
	}
	plainKey, err := util.GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}
	hash := sha256.Sum256([]byte(plainKey))
	newKey := models.APIKey{
		ID:          util.MustUUID(),
		UserID:      expired.UserID,
		Name:        expired.Name,
		KeyHash:     hex.EncodeToString(hash[:]),
		Permissions: expired.Permissions,
		ExpiresAt:   time.Now().Add(duration),
	}
	if err := s.db.Create(&newKey).Error; err != nil {
		return nil, "", err
	}
	return &newKey, plainKey, nil
}

var allowedPermissions = map[string]struct{}{
	"deposit":  {},
	"transfer": {},
	"read":     {},
}

func validatePermissions(perms []string) error {
	for _, p := range perms {
		if _, ok := allowedPermissions[util.NormalizePermission(p)]; !ok {
			return errors.New("invalid permission: " + p)
		}
	}
	return nil
}
