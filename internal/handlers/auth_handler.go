package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/CyberwizD/Wallet-Service/internal/auth"
	"github.com/CyberwizD/Wallet-Service/internal/config"
	"github.com/CyberwizD/Wallet-Service/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// AuthHandler manages Google auth endpoints.
type AuthHandler struct {
	cfg         config.Config
	userService *services.UserService
	oauthConfig *oauth2.Config
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(cfg config.Config, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		cfg:         cfg,
		userService: userService,
		oauthConfig: auth.NewGoogleOAuth(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL),
	}
}

// StartGoogleAuth redirects to Google's OAuth consent screen.
func (h *AuthHandler) StartGoogleAuth(c *gin.Context) {
	url := h.oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

// GoogleCallback exchanges the code, upserts the user, and returns a JWT.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}
	token, err := auth.ExchangeCode(context.Background(), h.oauthConfig, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot exchange code", "details": err.Error()})
		return
	}
	userInfo, err := auth.FetchGoogleUser(context.Background(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot fetch google user", "details": err.Error()})
		return
	}
	user, err := h.userService.UpsertGoogleUser(userInfo.Email, userInfo.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to persist user", "details": err.Error()})
		return
	}
	jwtToken, err := auth.GenerateToken(user.ID, user.Email, h.cfg.JWTSecret, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate jwt", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":        jwtToken,
		"user":         gin.H{"email": user.Email, "name": user.Name, "id": user.ID},
		"wallet":       gin.H{"number": user.Wallet.Number, "balance": user.Wallet.Balance},
		"expires_in_s": int64((24 * time.Hour).Seconds()),
	})
}
