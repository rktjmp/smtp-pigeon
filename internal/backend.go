package internal

import (
	"github.com/emersion/go-smtp"
)

// Permissive backend performs no authentication checks
type Permissive struct {
	config *Config
}

// NewBackend creates a New SMTP Pigeon Backend
func NewBackend(config *Config) *Permissive {
	be := Permissive{config: config}

	return &be
}

// Login creates a new session, any username or password is accepted
func (backend *Permissive) Login(_state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return makeSession(backend), nil
}

// AnonymousLogin creates a new session
func (backend *Permissive) AnonymousLogin(_state *smtp.ConnectionState) (smtp.Session, error) {
	return makeSession(backend), nil
}

func makeSession(backend *Permissive) *Session {
	session := newSession(backend.config)
	return session
}
