package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/CyberwizD/Wallet-Service/internal/config"
	"github.com/CyberwizD/Wallet-Service/internal/database"
	"github.com/CyberwizD/Wallet-Service/internal/models"
	"github.com/CyberwizD/Wallet-Service/internal/server"
)

func main() {
	// Load .env for local development; ignore if missing
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not loaded: %v", err)
	}

	cfg := config.Load()

	db := database.Connect(cfg.DBURL)
	if err := db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.APIKey{}, &models.Transaction{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	r := server.SetupRouter(cfg, db)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
