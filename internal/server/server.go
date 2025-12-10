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

	// Lightweight Swagger UI backed by docs/swagger.yaml
	r.StaticFile("/swagger.yaml", "docs/swagger.yaml")
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `
<!DOCTYPE html>
<html>
<head>
  <title>Wallet Service API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>body { margin: 0; padding: 0; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({
        url: '/swagger.yaml',
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>
		`)
	})

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
