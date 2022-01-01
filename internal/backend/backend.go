package backend

import (
	"github.com/emersion/go-smtp"
	"github.com/rktjmp/smtp-pigeon/internal/config"
	"github.com/rktjmp/smtp-pigeon/internal/session"
)

// Permissive backend performs no authentication checks
type Permissive struct {
	config *config.Config
}

// NewBackend creates a New SMTP Pigeon Backend
func NewBackend(config *config.Config) *Permissive {
	be := Permissive{config: config}

	return &be
}

// Login creates a new session, any username or password is accepted
func (backend *Permissive) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return backend.AnonymousLogin(state)
}

// AnonymousLogin creates a new session
func (backend *Permissive) AnonymousLogin(_state *smtp.ConnectionState) (smtp.Session, error) {
	session := session.NewSession(backend.config)
	return session, nil
}
