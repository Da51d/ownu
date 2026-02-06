package auth

import (
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// SessionStore provides in-memory storage for WebAuthn sessions
// In production, consider using Redis for distributed deployments
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionData
}

// SessionData holds WebAuthn session data
type SessionData struct {
	Session   *webauthn.SessionData
	Username  string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*SessionData),
	}
	// Start cleanup goroutine
	go store.cleanup()
	return store
}

// Save stores a session
func (s *SessionStore) Save(key string, session *webauthn.SessionData, username string, userID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[key] = &SessionData{
		Session:   session,
		Username:  username,
		UserID:    userID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
}

// Get retrieves a session
func (s *SessionStore) Get(key string) (*SessionData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.sessions[key]
	if !ok || time.Now().After(data.ExpiresAt) {
		return nil, false
	}
	return data, true
}

// Delete removes a session
func (s *SessionStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, key)
}

// cleanup periodically removes expired sessions
func (s *SessionStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, data := range s.sessions {
			if now.After(data.ExpiresAt) {
				delete(s.sessions, key)
			}
		}
		s.mu.Unlock()
	}
}
