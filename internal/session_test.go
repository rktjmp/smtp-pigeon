package internal

import (
	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
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

func TestDataBasicSettings(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sets json content type
		assert.Equal("application/json", r.Header.Get("Content-Type"))
		// posts
		assert.Equal("POST", r.Method, "POST method")
		// body as expected
		body, _ := ioutil.ReadAll(r.Body)
		assert.Equal("test@host", string(body), "POST method")
	}))
	defer server.Close()

	config, _ := NewConfig(server.URL, []string{}, "{{.From}}", false)
	session := newSession(config)
	session.Mail("test@host", smtp.MailOptions{})
	session.Rcpt("to@host")
	session.Data(strings.NewReader("my data\nlines"))
}

func TestDataComplexSettings(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// can over ride content type
		assert.Equal("text/plain", r.Header.Get("Content-Type"))
		// can set more than one c ustom headers
		assert.Equal("my-node", r.Header.Get("NodeID"))
		assert.Equal("POST", r.Method, "POST method")
		body, _ := ioutil.ReadAll(r.Body)
		assert.Equal("test@host", string(body), "POST method")
	}))
	defer server.Close()

	config, _ := NewConfig(server.URL, []string{"NodeID: my-node", "Content-Type: text/plain"}, "{{.From}}", false)
	session := newSession(config)
	session.Mail("test@host", smtp.MailOptions{})
	session.Rcpt("to@host")
	session.Data(strings.NewReader("my data\nlines"))
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

func TestNewSession(t *testing.T) {
	assert := assert.New(t)
	config := &Config{}
	session := newSession(config)
	assert.NotEqual("", session.id, "generates id")
}

func TestCheckTemplateExecute(t *testing.T) {
	assert := assert.New(t)

	// No error on a good template
	var template, err = template.New("test").Parse("{{.ID}}")
	assert.Nil(err)
	err = CheckTemplateExecute(template)
	assert.Nil(err)

	// Misformed templates are
	template, err = template.New("test").Parse("{{.BadKey}}")
	assert.Nil(err)
	err = CheckTemplateExecute(template)
	assert.NotNil(err)
}

func TestPreparePostBody(t *testing.T) {
	assert := assert.New(t)
	session := &Session{
		id:   "test-id",
		from: "from@host",
		to:   []string{"to@host", "too@host"},
		data: "data string",
	}
	var template, err = template.New("test").Parse(`{{.ID}},{{.From}},{{range $i, $e := .To}}{{if $i}},{{end}}{{$e}}{{end}},{{.Data}}`)
	bytes, err := preparePostBody(session, template)
	assert.Nil(err)
	assert.NotNil(bytes)
	assert.Equal("test-id,from@host,to@host,too@host,data string", bytes.String(), "apply template")

	template, err = template.New("test").Parse("{{.Bad}}")
	bytes, err = preparePostBody(session, template)
	assert.NotNil(err)
}
