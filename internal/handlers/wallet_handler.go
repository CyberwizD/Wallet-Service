package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/CyberwizD/Wallet-Service/internal/middleware"
	"github.com/CyberwizD/Wallet-Service/internal/services"

	"github.com/gin-gonic/gin"
)

// WalletHandler exposes wallet endpoints.
type WalletHandler struct {
	walletService *services.WalletService
	paystack      *services.PaystackService
}

// NewWalletHandler constructs a WalletHandler.
func NewWalletHandler(walletService *services.WalletService, paystack *services.PaystackService) *WalletHandler {
	return &WalletHandler{walletService: walletService, paystack: paystack}
}

type depositRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

// Deposit starts a Paystack deposit flow.
func (h *WalletHandler) Deposit(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	var req depositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	ref, authURL, err := h.walletService.InitiateDeposit(user, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reference": ref, "authorization_url": authURL})
}

// PaystackWebhook handles Paystack transaction events (must remain idempotent).
func (h *WalletHandler) PaystackWebhook(c *gin.Context) {
	signature := c.GetHeader("x-paystack-signature")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}
	if !h.paystack.VerifySignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}
	var event paystackWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if event.Data.Reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing reference"})
		return
	}
	if err := h.walletService.ApplyDepositWebhook(event.Data.Reference, event.Data.Status, body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}

type paystackWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
	} `json:"data"`
}

// DepositStatus returns the status of a deposit reference without crediting wallets.
func (h *WalletHandler) DepositStatus(c *gin.Context) {
	ref := c.Param("reference")
	tx, err := h.walletService.DepositStatus(ref)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reference not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reference": tx.Reference,
		"status":    tx.Status,
		"amount":    tx.Amount,
	})
}

// Balance returns the caller's wallet balance.
func (h *WalletHandler) Balance(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	bal, err := h.walletService.Balance(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": bal, "wallet_number": user.Wallet.Number})
}

type transferRequest struct {
	WalletNumber string `json:"wallet_number" binding:"required"`
	Amount       int64  `json:"amount" binding:"required"`
}

// Transfer moves funds to another wallet.
func (h *WalletHandler) Transfer(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	if err := h.walletService.Transfer(user, req.WalletNumber, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Transfer completed"})
}

// Transactions lists wallet activity.
func (h *WalletHandler) Transactions(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	txs, err := h.walletService.Transactions(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := make([]gin.H, 0, len(txs))
	for _, t := range txs {
		resp = append(resp, gin.H{
			"type":       t.Type,
			"amount":     t.Amount,
			"status":     t.Status,
			"reference":  t.Reference,
			"created_at": t.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, resp)
}
