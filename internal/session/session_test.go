package session

import (
	"github.com/emersion/go-smtp"
	"github.com/rktjmp/smtp-pigeon/internal/config"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"
)

func TestSessionMailSetsFrom(t *testing.T) {
	assert := assert.New(t)

	session := &Session{
		id: "test-id",
	}
	assert.Equal("", session.from, "from defaults to blank")
	session.Mail("me@host", smtp.MailOptions{})
	assert.Equal("me@host", session.from, "from value was set")
}

func TestSessionRcptAppentsTo(t *testing.T) {
	assert := assert.New(t)

	session := &Session{
		id: "test-id",
	}

	assert.Equal(0, len(session.to), "to is empty by default")
	session.Rcpt("a@host")
	assert.Equal([]string{"a@host"}, session.to, "Rcpt appends to list")
	session.Rcpt("b@host")
	assert.Equal([]string{"a@host", "b@host"}, session.to, "Rcpt appends to list")
}

func TestDataWithBadInput(t *testing.T) {
	assert := assert.New(t)
	session := &Session{
		id: "test-id",
	}
	err := session.Data(strings.NewReader(""))
	assert.NotNil(err)
}

func TestData(t *testing.T) {
	assert := assert.New(t)

	// data needs to hit a server in the end
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.True(len(body) > 0)
	}))
	defer server.Close()

	// this is pretty brittle/coupled test
	cfg, _ := config.NewConfig(server.URL, []string{}, "{{.ID}}", false)
	data := `Subject: ON MY WAY
From: Gordon Freeman <freeman@materials.blackmesa.com>
To: Eli Vance <vance@materials.blackmesa.com>
Cc: Issac Kleiner <kleiner@materials.blackmesa.com>

hey guys running L8 2DAY
on the tram now`

	session := NewSession(cfg)
	session.Mail("freeman@mailhub.bm.net", smtp.MailOptions{})
	session.Rcpt("vance@mailhub.bm.net")
	session.Rcpt("kleiner@mailhub.bm.net")

	// finally hit data
	err := session.Data(strings.NewReader(data))
	assert.Nil(err)
}

func TestReset(t *testing.T) {
	assert := assert.New(t)

	session := &Session{
		id:   "test-id",
		from: "from@host",
	}
	session.Reset()
	assert.NotEqual("test-id", session.id, "id is refreshed")
}

func TestLogout(t *testing.T) {
	assert := assert.New(t)

	session := &Session{
		id:   "test-id",
		from: "from@host",
	}
	session.Logout()
	assert.True(session.ended)
}

func TestNewSession(t *testing.T) {
	assert := assert.New(t)
	config := &config.Config{}
	session := NewSession(config)
	assert.NotEqual("", session.id, "generates id")
}

func TestTemplateData(t *testing.T) {
	assert := assert.New(t)

	message := &mail.Message{}
	s := &Session{
		id:        "my-id",
		timestamp: time.Now(),
		from:      "me",
		to:        []string{"you"},
		data:      "data\nmy-message",
		body:      "my-message",
		message:   message,
	}
	td := s.TemplateData()

	assert.NotNil(td)
	assert.Equal(td.ID, "my-id")
	assert.Equal(td.Timestamp, s.timestamp)
	assert.Equal(td.Sender, "me")
	assert.Equal(td.Recipients, []string{"you"})
	assert.Equal(td.Data, "data\nmy-message")
	assert.Equal(td.Body, "my-message")
	assert.IsType(td.Header, mail.Header{})
}
