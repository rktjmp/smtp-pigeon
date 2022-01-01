package session

import (
	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"github.com/rktjmp/smtp-pigeon/internal/config"
	"github.com/rktjmp/smtp-pigeon/internal/dispatch"
	"io"
	"log"
	"net/mail"
	"strings"
	"time"
)

// Session holds data for each connected SMTP connection
type Session struct {
	config    *config.Config
	ended     bool
	sent      bool
	id        string
	timestamp time.Time
	from      string
	to        []string
	data      string
	message   *mail.Message
	body      string
}

// NewSession creates a fresh session with a generated UUID and timestamp
func NewSession(config *config.Config) *Session {
	id := uuid.New().String()
	s := &Session{
		id:        id,
		config:    config,
		timestamp: time.Now(),
	}
	log.Printf("%v: New session", s.id)
	return s
}

// Mail is called on the MAIL SMTP command, it only stores the from address
func (s *Session) Mail(from string, _ smtp.MailOptions) error {
	log.Printf("%v: MAIL: %v", s.id, from)
	s.from = from
	return nil
}

// Rcpt is called on the RCPT SMTP command, it stores the to address. It may be
// called multiple times in one session.
func (s *Session) Rcpt(to string) error {
	log.Printf("%v: RCPT: %v", s.id, to)
	s.to = append(s.to, to)
	return nil
}

// Data is called on the DATA SMTP command. It generally contains the "message"
// but also any other informational headers such as the mailer client and
// subject. This is seen as the "finished" command for each session, so it
// triggers the HTTP post, though technically Reset and Logout wil be called
// afterwards.
func (s *Session) Data(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("%v: %v", s.id, err)
	}
	log.Printf("%v: DATA: [redacted (%d bytes)]", s.id, len(b))

	// store the raw data and generate a mail.Message too
	s.data = string(b)
	s.message, err = mail.ReadMessage(strings.NewReader(s.data))
	if err != nil {
		// There is a world where reading the message fails, but we could still send
		// raw data but not sure what the best user interface for that is. For now we
		// just die.
		log.Printf("%v: mail.ReadMessage failed (misbehaving sender?), refusing to post: %v", s.id, err)
		return err
	}
	b, _ = io.ReadAll(s.message.Body)
	s.body = string(b)

	// DATA dispatch POST request
	endpoint := &dispatch.Endpoint{
		URL:     s.config.URL,
		Headers: s.config.Headers,
	}

	templateData := s.TemplateData()

	status, err := dispatch.POST(endpoint, s.config.Template, templateData)
	if err != nil {
		log.Printf("%v: POST failed: %v", s.id, err)
	} else {
		s.sent = true
		log.Printf("%v: POST returned status: %v", s.id, status)
	}

	return err
}

var zeroSession = &Session{}

// Reset is called on the RSET SMTP command, or after a successful DATA command
// It will log whether the current session did or not make a post request.
func (s *Session) Reset() {
	if s.sent {
		log.Printf("%v: Session reset after POST", s.id)
	} else {
		log.Printf("%v: Session reset without POST", s.id)
	}
	new := NewSession(s.config)
	*s = *new
}

// Logout is called when a connection is terminated.
// It will log whether the current session did or not make a post request.
func (s *Session) Logout() error {
	if s.sent {
		log.Printf("%v: Session logout after POST", s.id)
	} else {
		log.Printf("%v: Session logout without POST", s.id)
	}
	s.ended = true
	return nil
}

func (s *Session) TemplateData() *dispatch.TemplateData {
	return &dispatch.TemplateData{
		ID:         s.id,
		Timestamp:  s.timestamp,
		Sender:     s.from,
		Recipients: s.to,
		Data:       s.data,
		Body:       s.body,
		Header:     s.message.Header,
	}
}
