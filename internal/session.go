package internal

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

// Session holds data for each connected SMTP connection
type Session struct {
	config     *Config
	sent       bool
	id         string
	receivedAt time.Time
	from       string
	to         []string
	data       string
}

// we will use this struct in the template
type templateData struct {
	ID         string
	ReceivedAt string
	From       string
	To         []string
	Data       string
}

// Mail is called on the MAIL SMTP command, it only stores the from address
func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	log.Printf("%v: Mail from: %v", s.id, from)
	s.from = from
	s.receivedAt = time.Now().UTC()
	return nil
}

// Rcpt is called on the RCPT SMTP command, it stores the to address. It may be
// called multiple times in one session.
func (s *Session) Rcpt(to string) error {
	log.Printf("%v: Rcpt to: %v", s.id, to)
	s.to = append(s.to, to)
	return nil
}

// Data is called on the DATA SMTP command. It generally contains the "message"
// but also any other information such as the mailer client and subject.
// This is seen as the "finished" command for each session, so it triggers the
// HTTP post, though technically Reset and Logout wil be called afterwards.
func (s *Session) Data(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("%v: %v", s.id, err)
	}

	s.data = string(b)
	log.Printf("%v: Data: [redacted (%d bytes)]", s.id, len(s.data))

	var jsonBuffer *bytes.Buffer
	jsonBuffer, err = preparePostBody(s, s.config.template)
	if err != nil {
		log.Printf("%v: Could not execute template, unable to make POST: %v", s.id, err)
		return errors.New("Template error")
	}

	if s.config.verbose {
		log.Printf("%v: POST content: %v", s.id, jsonBuffer.String())
	}

	log.Printf("%v: Performing HTTP POST to %v", s.id, s.config.url)
	resp, err := performPOSTRequest(s.config.url, s.config.headers, jsonBuffer)
	if err != nil {
		log.Printf("%v: HTTP Post encountered an error: %v", s.id, err)
		return errors.New("HTTP POST failed")
	}

	defer resp.Body.Close()
	log.Printf("%v: HTTP POST responded with %v", s.id, resp.StatusCode)
	s.sent = true

	return nil
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
	new := newSession(s.config)
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
	return nil
}

func newSession(config *Config) *Session {
	id := uuid.New().String()
	log.Printf("%v: New session", id)
	return &Session{id: id, config: config}
}

// CheckTemplateExecute generates a fake session and attempts to apply it to the given template.
// This is used to pre-flight check a user's template will succeed once mail comes in.
func CheckTemplateExecute(tmpl *template.Template) error {
	s := &Session{
		id:         "my-id",
		from:       "test@address",
		to:         []string{"to1@address", "to2@address"},
		receivedAt: time.Now().UTC(),
		data:       "This is my message",
	}
	_, err := preparePostBody(s, tmpl)
	return err
}

func sessionToTemplateData(s *Session) *templateData {
	// Perhaps overly pedantically, we want all content sent by the
	// client in the "DATA" command available under "Data". Data is
	// a function on the Session struct, so we can't use the same name.
	// Instead we will copy our data over to a more "friendly" struct that really
	// just renames that one field.
	// This could be done with funcMap on the template, but then you're defining that
	// else where and still kind of leaking the Session abstraction.
	// This will do for now.
	return &templateData{
		ID:         s.id,
		ReceivedAt: s.receivedAt.Format(time.RFC3339),
		From:       s.from,
		To:         s.to,
		Data:       s.data,
	}
}

// Convert sesson to templateData and pass through template.
// returns
func preparePostBody(s *Session, tmpl *template.Template) (*bytes.Buffer, error) {
	data := sessionToTemplateData(s)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func performPOSTRequest(url string, headers [][]string, body *bytes.Buffer) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("Unable to create HTTP request: %v", err)
	}

	// set content type first, this can be overridden by a --header if desired
	req.Header.Set("Content-Type", "application/json")
	for _, pair := range headers {
		req.Header.Set(pair[0], pair[1])
	}

	return client.Do(req)
}
