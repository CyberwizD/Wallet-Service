package server

import (
	"github.com/CyberwizD/Wallet-Service/internal/config"
	"github.com/CyberwizD/Wallet-Service/internal/handlers"
	"github.com/CyberwizD/Wallet-Service/internal/middleware"
	"github.com/CyberwizD/Wallet-Service/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter wires dependencies and routes.
func SetupRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	paystack := services.NewPaystackService(cfg.PaystackSecret, cfg.PaystackBaseURL)
	userService := services.NewUserService(db)
	walletService := services.NewWalletService(db, paystack)
	keyService := services.NewAPIKeyService(db)

	authHandler := handlers.NewAuthHandler(cfg, userService)
	keyHandler := handlers.NewKeyHandler(keyService)
	walletHandler := handlers.NewWalletHandler(walletService, paystack)

	r := gin.Default()
	r.GET("/auth/google", authHandler.StartGoogleAuth)
	r.GET("/auth/google/callback", authHandler.GoogleCallback)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(db, cfg.JWTSecret))
	{
		protected.POST("/keys/create", keyHandler.CreateKey)
		protected.POST("/keys/rollover", keyHandler.RolloverKey)

		protected.POST("/wallet/deposit", middleware.RequirePermission("deposit"), walletHandler.Deposit)
		protected.GET("/wallet/deposit/:reference/status", middleware.RequirePermission("read"), walletHandler.DepositStatus)
		protected.GET("/wallet/balance", middleware.RequirePermission("read"), walletHandler.Balance)
		protected.POST("/wallet/transfer", middleware.RequirePermission("transfer"), walletHandler.Transfer)
		protected.GET("/wallet/transactions", middleware.RequirePermission("read"), walletHandler.Transactions)
	}

	r.POST("/wallet/paystack/webhook", walletHandler.PaystackWebhook)
	return r
}
