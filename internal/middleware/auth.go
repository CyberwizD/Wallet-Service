package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/CyberwizD/Wallet-Service/internal/auth"
	"github.com/CyberwizD/Wallet-Service/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type contextKey string

const (
	contextUserKey   contextKey = "currentUser"
	contextAPIKeyKey contextKey = "currentAPIKey"
)

// AuthMiddleware populates the request context with either a JWT user or an API key principal.
func AuthMiddleware(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tryJWT(c, db, jwtSecret) {
			c.Next()
			return
		}
		if tryAPIKey(c, db) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
	}
}

// RequirePermission asserts the current principal has the given permission.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if u := GetUser(c); u != nil {
			c.Next()
			return
		}
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		perms := strings.Split(apiKey.Permissions, ",")
		want := strings.ToLower(permission)
		for _, p := range perms {
			if strings.TrimSpace(strings.ToLower(p)) == want {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	}
}

// GetUser returns the JWT-authenticated user if present.
func GetUser(c *gin.Context) *models.User {
	if val, exists := c.Get(string(contextUserKey)); exists {
		if user, ok := val.(*models.User); ok {
			return user
		}
	}
	return nil
}

// GetAPIKey returns the API key principal if present.
func GetAPIKey(c *gin.Context) *models.APIKey {
	if val, exists := c.Get(string(contextAPIKeyKey)); exists {
		if key, ok := val.(*models.APIKey); ok {
			return key
		}
	}
	return nil
}

func tryJWT(c *gin.Context, db *gorm.DB, jwtSecret string) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return false
	}
	claims, err := auth.ParseToken(parts[1], jwtSecret)
	if err != nil {
		return false
	}
	var user models.User
	if err := db.Preload("Wallet").First(&user, "id = ?", claims.UserID).Error; err != nil {
		return false
	}
	c.Set(string(contextUserKey), &user)
	return true
}

func tryAPIKey(c *gin.Context, db *gorm.DB) bool {
	key := c.GetHeader("x-api-key")
	if key == "" {
		return false
	}
	hash := sha256.Sum256([]byte(key))
	var record models.APIKey
	if err := db.First(&record, "key_hash = ?", hex.EncodeToString(hash[:])).Error; err != nil {
		return false
	}
	if record.Revoked || time.Now().After(record.ExpiresAt) {
		return false
	}
	var user models.User
	if err := db.Preload("Wallet").First(&user, "id = ?", record.UserID).Error; err != nil {
		return false
	}
	now := time.Now()
	record.LastUsedAt = &now
	_ = db.Model(&record).Update("last_used_at", now).Error
	c.Set(string(contextUserKey), &user)
	c.Set(string(contextAPIKeyKey), &record)
	return true
}
