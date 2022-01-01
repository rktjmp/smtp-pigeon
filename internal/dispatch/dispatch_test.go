package dispatch

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"testing"
	"text/template"
	"time"
)

func makeTemplateData() *TemplateData {
	return &TemplateData{
		ID:         "constant-id",
		Timestamp:  time.Now(),
		Sender:     "me@host",
		Recipients: []string{"you@host", "them@host"},
		Data:       "My message",
		Header:     mail.Header{},
		Body:       "My message",
	}
}

func TestPOSTNoHeaders(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sets json content type
		assert.Equal("application/json", r.Header.Get("Content-Type"))
		// posts
		assert.Equal("POST", r.Method, "POST method")
		// body as expected
		body, _ := io.ReadAll(r.Body)
		assert.Equal("constant-id", string(body), "POST method")
	}))
	defer server.Close()

	ep := &Endpoint{
		URL:     server.URL,
		Headers: [][2]string{},
	}
	data := makeTemplateData()

	tmpl, err := template.New("test").Parse("{{.ID}}")
	assert.Nil(err)
	status, err := POST(ep, tmpl, data)
	assert.NotNil(status)
}

func TestPOSTCustomHeaders(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sets json content type
		assert.Equal("text/plain", r.Header.Get("Content-Type"))
		assert.Equal("my-node", r.Header.Get("NodeID"))
	}))
	defer server.Close()

	ep := &Endpoint{
		URL: server.URL,
		Headers: [][2]string{
			[2]string{"Content-Type", "text/plain"},
			[2]string{"NodeID", "my-node"},
		},
	}

	data := makeTemplateData()

	tmpl, err := template.New("test").Parse("{{.ID}}")
	assert.Nil(err)
	status, err := POST(ep, tmpl, data)
	assert.NotNil(status)
}

func TestPOSTServerError(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	server.Close()

	ep := &Endpoint{
		URL:     server.URL,
		Headers: [][2]string{},
	}
	data := makeTemplateData()

	tmpl, err := template.New("test").Parse("{{.ID}}")
	assert.Nil(err)
	status, err := POST(ep, tmpl, data)
	assert.Equal(0, status)
	assert.NotNil(err)
}

func TestPOSTTemplateExecuteError(t *testing.T) {
	assert := assert.New(t)

	ep := &Endpoint{
		URL:     "anything",
		Headers: [][2]string{},
	}
	data := makeTemplateData()

	tmpl, err := template.New("test").Parse("{{.IDs}}")
	assert.Nil(err)
	status, err := POST(ep, tmpl, data)
	assert.Equal(0, status)
	assert.NotNil(err)
}
