package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	// HSTS settings
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// CSP settings
	CSPDirectives map[string][]string

	// Rate limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Other settings
	FrameOptions     string
	ContentTypeNoSniff bool
	XSSProtection    bool
	ReferrerPolicy   string
}

// DefaultSecurityConfig returns secure defaults
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
		CSPDirectives: map[string][]string{
			"default-src": {"'self'"},
			"script-src":  {"'self'"},
			"style-src":   {"'self'", "'unsafe-inline'"}, // Allow inline styles for React
			"img-src":     {"'self'", "data:", "https:"},
			"font-src":    {"'self'"},
			"connect-src": {"'self'", "https://*.plaid.com"}, // Allow Plaid connections
			"frame-src":   {"'self'", "https://*.plaid.com"}, // Allow Plaid Link iframe
			"object-src":  {"'none'"},
			"base-uri":    {"'self'"},
			"form-action": {"'self'"},
		},
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		FrameOptions:       "DENY",
		ContentTypeNoSniff: true,
		XSSProtection:      true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// SecurityHeaders middleware adds security headers to responses
func SecurityHeaders(config SecurityConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()

			// HSTS (only on HTTPS)
			if c.Scheme() == "https" || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				hstsValue := buildHSTSValue(config)
				res.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Content Security Policy
			cspValue := buildCSPValue(config.CSPDirectives)
			res.Header().Set("Content-Security-Policy", cspValue)

			// X-Frame-Options
			if config.FrameOptions != "" {
				res.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-Content-Type-Options
			if config.ContentTypeNoSniff {
				res.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// X-XSS-Protection (legacy but still useful for older browsers)
			if config.XSSProtection {
				res.Header().Set("X-XSS-Protection", "1; mode=block")
			}

			// Referrer-Policy
			if config.ReferrerPolicy != "" {
				res.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			// Permissions-Policy (formerly Feature-Policy)
			res.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Cache control for API responses
			if strings.HasPrefix(c.Path(), "/api/") {
				res.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				res.Header().Set("Pragma", "no-cache")
			}

			return next(c)
		}
	}
}

func buildHSTSValue(config SecurityConfig) string {
	value := "max-age=" + string(rune(config.HSTSMaxAge))
	if config.HSTSIncludeSubdomains {
		value += "; includeSubDomains"
	}
	if config.HSTSPreload {
		value += "; preload"
	}
	return value
}

func buildCSPValue(directives map[string][]string) string {
	var parts []string
	for directive, values := range directives {
		parts = append(parts, directive+" "+strings.Join(values, " "))
	}
	return strings.Join(parts, "; ")
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this key
	requests := rl.requests[key]

	// Filter to only requests within the window
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if over limit
	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	// Add this request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	windowStart := time.Now().Add(-rl.window)
	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use IP address as the key (in production, consider user ID for authenticated requests)
			key := c.RealIP()

			if !rl.Allow(key) {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if request ID already exists (from reverse proxy)
			reqID := c.Request().Header.Get("X-Request-ID")
			if reqID == "" {
				// Generate new request ID
				bytes := make([]byte, 16)
				rand.Read(bytes)
				reqID = hex.EncodeToString(bytes)
			}

			// Set in response header
			c.Response().Header().Set("X-Request-ID", reqID)

			// Store in context for logging
			c.Set("request_id", reqID)

			return next(c)
		}
	}
}

// SecureRedirectMiddleware redirects HTTP to HTTPS
func SecureRedirectMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if already HTTPS
			if c.Scheme() == "https" || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				return next(c)
			}

			// Skip for health checks
			if c.Path() == "/health" {
				return next(c)
			}

			// Redirect to HTTPS
			httpsURL := "https://" + c.Request().Host + c.Request().RequestURI
			return c.Redirect(http.StatusMovedPermanently, httpsURL)
		}
	}
}
