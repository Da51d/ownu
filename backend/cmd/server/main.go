package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ownu/ownu/internal/api"
	"github.com/ownu/ownu/internal/auth"
	"github.com/ownu/ownu/internal/config"
	"github.com/ownu/ownu/internal/plaid"
	"github.com/ownu/ownu/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	db, err := repository.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	plaidRepo := repository.NewPlaidRepository(db)

	// Initialize WebAuthn
	webauthn, err := auth.NewWebAuthnService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn: %v", err)
	}

	// Initialize Plaid (optional - will be nil if not configured)
	var plaidService *plaid.Service
	if plaid.IsConfigured(cfg) {
		plaidService, err = plaid.NewService(cfg)
		if err != nil {
			log.Printf("Warning: Plaid integration disabled: %v", err)
		} else {
			log.Println("Plaid integration enabled")
		}
	} else {
		log.Println("Plaid integration not configured (PLAID_CLIENT_ID and PLAID_SECRET required)")
	}

	// Create and start server
	server := api.NewServer(cfg, userRepo, accountRepo, plaidRepo, webauthn, plaidService)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		_ = shutdownCtx
	}()

	// Start server
	address := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Starting server on %s", address)
	if err := server.Start(address); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}
