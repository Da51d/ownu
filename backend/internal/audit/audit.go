package audit

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of audit event
type EventType string

const (
	// Authentication events
	EventAuthLoginAttempt   EventType = "auth.login.attempt"
	EventAuthLoginSuccess   EventType = "auth.login.success"
	EventAuthLoginFailure   EventType = "auth.login.failure"
	EventAuthLogout         EventType = "auth.logout"
	EventAuthRegister       EventType = "auth.register"
	EventAuthRecovery       EventType = "auth.recovery"

	// Data access events
	EventDataRead   EventType = "data.read"
	EventDataCreate EventType = "data.create"
	EventDataUpdate EventType = "data.update"
	EventDataDelete EventType = "data.delete"
	EventDataExport EventType = "data.export"

	// Integration events
	EventPlaidConnect    EventType = "plaid.connect"
	EventPlaidDisconnect EventType = "plaid.disconnect"
	EventPlaidSync       EventType = "plaid.sync"

	// Security events
	EventSecurityAuthzFailure EventType = "security.authz.failure"
	EventSecurityRateLimit    EventType = "security.rate_limit"
	EventSecuritySuspicious   EventType = "security.suspicious"

	// System events
	EventSystemStartup  EventType = "system.startup"
	EventSystemShutdown EventType = "system.shutdown"
	EventSystemConfig   EventType = "system.config_change"
)

// Outcome represents the result of an action
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
	OutcomeDenied  Outcome = "denied"
)

// Event represents an audit log event
type Event struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Type      EventType         `json:"type"`
	UserID    string            `json:"user_id,omitempty"`
	SessionID string            `json:"session_id,omitempty"`
	Resource  string            `json:"resource,omitempty"`
	Action    string            `json:"action,omitempty"`
	Outcome   Outcome           `json:"outcome"`
	Reason    string            `json:"reason,omitempty"`
	IP        string            `json:"ip,omitempty"`
	UserAgent string            `json:"user_agent,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

// Logger is the audit logger interface
type Logger interface {
	Log(ctx context.Context, event Event)
	Query(ctx context.Context, filter QueryFilter) ([]Event, error)
	Close() error
}

// QueryFilter for querying audit logs
type QueryFilter struct {
	UserID    string
	Type      EventType
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// FileLogger writes audit logs to a file
type FileLogger struct {
	file   *os.File
	mu     sync.Mutex
	events []Event // In-memory buffer for queries (simple implementation)
}

// NewFileLogger creates a new file-based audit logger
func NewFileLogger(path string) (*FileLogger, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &FileLogger{
		file:   file,
		events: make([]Event, 0),
	}, nil
}

// Log writes an audit event
func (l *FileLogger) Log(ctx context.Context, event Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Set defaults
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Store in memory for queries (in production, use database)
	l.events = append(l.events, event)

	// Keep only last 10000 events in memory
	if len(l.events) > 10000 {
		l.events = l.events[len(l.events)-10000:]
	}

	// Write to file
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal audit event: %v", err)
		return
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		log.Printf("Failed to write audit event: %v", err)
	}
}

// Query retrieves audit events matching the filter
func (l *FileLogger) Query(ctx context.Context, filter QueryFilter) ([]Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var results []Event
	for _, event := range l.events {
		// Apply filters
		if filter.UserID != "" && event.UserID != filter.UserID {
			continue
		}
		if filter.Type != "" && event.Type != filter.Type {
			continue
		}
		if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
			continue
		}

		results = append(results, event)
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(results) {
		results = results[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

// Close closes the audit logger
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// NullLogger is a no-op logger for testing
type NullLogger struct{}

func (l *NullLogger) Log(ctx context.Context, event Event)                      {}
func (l *NullLogger) Query(ctx context.Context, filter QueryFilter) ([]Event, error) {
	return nil, nil
}
func (l *NullLogger) Close() error { return nil }

// Helper functions for common events

// LogAuth logs an authentication event
func LogAuth(l Logger, ctx context.Context, eventType EventType, userID, ip, userAgent string, outcome Outcome, reason string) {
	l.Log(ctx, Event{
		Type:      eventType,
		UserID:    userID,
		Outcome:   outcome,
		Reason:    reason,
		IP:        ip,
		UserAgent: userAgent,
	})
}

// LogDataAccess logs a data access event
func LogDataAccess(l Logger, ctx context.Context, eventType EventType, userID, resource, action string, outcome Outcome) {
	l.Log(ctx, Event{
		Type:     eventType,
		UserID:   userID,
		Resource: resource,
		Action:   action,
		Outcome:  outcome,
	})
}

// LogSecurityEvent logs a security event
func LogSecurityEvent(l Logger, ctx context.Context, eventType EventType, userID, ip string, details map[string]string) {
	l.Log(ctx, Event{
		Type:    eventType,
		UserID:  userID,
		IP:      ip,
		Outcome: OutcomeDenied,
		Details: details,
	})
}
