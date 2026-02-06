package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ownu/ownu/internal/auth"
	"github.com/ownu/ownu/internal/config"
	plaidSvc "github.com/ownu/ownu/internal/plaid"
	"github.com/ownu/ownu/internal/repository"
)

// Server holds the API server dependencies
type Server struct {
	echo        *echo.Echo
	config      *config.Config
	userRepo    *repository.UserRepository
	accountRepo *repository.AccountRepository
	plaidRepo   *repository.PlaidRepository
	webauthn    *auth.WebAuthnService
	plaid       *plaidSvc.Service
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, userRepo *repository.UserRepository, accountRepo *repository.AccountRepository, plaidRepo *repository.PlaidRepository, webauthn *auth.WebAuthnService, plaid *plaidSvc.Service) *Server {
	e := echo.New()
	e.HideBanner = true

	// Security middleware (apply first)
	e.Use(RequestIDMiddleware())
	e.Use(SecurityHeaders(DefaultSecurityConfig()))

	// Standard middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.WebAuthnRPOrigin},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Rate limiting
	rateLimiter := NewRateLimiter(100, time.Minute)
	e.Use(RateLimitMiddleware(rateLimiter))

	s := &Server{
		echo:        e,
		config:      cfg,
		userRepo:    userRepo,
		accountRepo: accountRepo,
		plaidRepo:   plaidRepo,
		webauthn:    webauthn,
		plaid:       plaid,
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up all API routes
func (s *Server) registerRoutes() {
	// Health check
	s.echo.GET("/health", s.healthCheck)

	// API v1
	v1 := s.echo.Group("/api/v1")

	// Auth routes
	authGroup := v1.Group("/auth")
	authGroup.POST("/register/begin", s.beginRegistration)
	authGroup.POST("/register/finish", s.finishRegistration)
	authGroup.POST("/login/begin", s.beginLogin)
	authGroup.POST("/login/finish", s.finishLogin)

	// Protected routes (require JWT)
	protected := v1.Group("")
	protected.Use(s.jwtMiddleware)

	// Accounts
	protected.GET("/accounts", s.listAccounts)
	protected.POST("/accounts", s.createAccount)
	protected.GET("/accounts/:id", s.getAccount)
	protected.PUT("/accounts/:id", s.updateAccount)
	protected.DELETE("/accounts/:id", s.deleteAccount)

	// Transactions
	protected.GET("/transactions", s.listTransactions)
	protected.POST("/transactions", s.createTransaction)
	protected.GET("/transactions/:id", s.getTransaction)
	protected.PUT("/transactions/:id", s.updateTransaction)
	protected.DELETE("/transactions/:id", s.deleteTransaction)

	// Categories
	protected.GET("/categories", s.listCategories)
	protected.POST("/categories", s.createCategory)
	protected.PUT("/categories/:id", s.updateCategory)
	protected.DELETE("/categories/:id", s.deleteCategory)

	// Import
	protected.POST("/import/csv", s.importCSV)
	protected.POST("/import/ofx", s.importOFX)
	protected.GET("/import/:id/preview", s.previewImport)
	protected.POST("/import/:id/confirm", s.confirmImport)

	// Reports
	protected.GET("/reports/spending", s.spendingReport)
	protected.GET("/reports/cashflow", s.cashflowReport)

	// Plaid integration (public status check)
	v1.GET("/plaid/status", s.plaidStatus)

	// Plaid protected routes
	protected.POST("/plaid/link-token", s.createLinkToken)
	protected.POST("/plaid/exchange-token", s.exchangePublicToken)
	protected.GET("/plaid/items", s.listPlaidItems)
	protected.GET("/plaid/items/:id", s.getPlaidItem)
	protected.DELETE("/plaid/items/:id", s.deletePlaidItem)
	protected.POST("/plaid/items/:id/sync", s.syncTransactions)

	// Privacy and data management (GDPR/CCPA compliance)
	protected.GET("/privacy/export", s.exportData)
	protected.GET("/privacy/export/csv", s.exportTransactionsCSV)
	protected.DELETE("/privacy/account", s.deleteUserAccount)
	protected.GET("/privacy/settings", s.getPrivacySettings)
	protected.GET("/privacy/consent", s.getConsentStatus)
}

// Start starts the HTTP server
func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.echo.Close()
}

// healthCheck returns server health status
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Placeholder handlers - to be implemented

func (s *Server) listTransactions(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) createTransaction(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) getTransaction(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) updateTransaction(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) deleteTransaction(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) listCategories(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) createCategory(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) updateCategory(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) deleteCategory(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) importCSV(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) importOFX(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) previewImport(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) confirmImport(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) spendingReport(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) cashflowReport(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}
