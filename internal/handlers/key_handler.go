package handlers

import (
	"net/http"

	"github.com/CyberwizD/Wallet-Service/internal/middleware"
	"github.com/CyberwizD/Wallet-Service/internal/services"

	"github.com/gin-gonic/gin"
)

// KeyHandler manages API key endpoints.
type KeyHandler struct {
	service *services.APIKeyService
}

// NewKeyHandler constructs a KeyHandler.
func NewKeyHandler(service *services.APIKeyService) *KeyHandler {
	return &KeyHandler{service: service}
}

type createKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
	Expiry      string   `json:"expiry" binding:"required"`
}

// CreateKey issues a new API key for a JWT-authenticated user.
func (h *KeyHandler) CreateKey(c *gin.Context) {
	if middleware.GetAPIKey(c) != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "API keys cannot create new keys"})
		return
	}
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user authentication required"})
		return
	}
	var req createKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	key, plain, err := h.service.CreateKey(user, req.Name, req.Permissions, req.Expiry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"api_key":     plain,
		"expires_at":  key.ExpiresAt,
		"permissions": key.Permissions,
	})
}

type rolloverKeyRequest struct {
	ExpiredKeyID string `json:"expired_key_id" binding:"required"`
	Expiry       string `json:"expiry" binding:"required"`
}

// RolloverKey issues a new key reusing permissions from an expired key.
func (h *KeyHandler) RolloverKey(c *gin.Context) {
	if middleware.GetAPIKey(c) != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "API keys cannot rollover keys"})
		return
	}
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user authentication required"})
		return
	}
	var req rolloverKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	key, plain, err := h.service.RolloverKey(user.ID, req.ExpiredKeyID, req.Expiry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"api_key":     plain,
		"expires_at":  key.ExpiresAt,
		"permissions": key.Permissions,
	})
}
