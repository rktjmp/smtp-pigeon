package internal

import (
	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHeaderStringsToPairs(t *testing.T) {
	assert := assert.New(t)

	var err error

	args := []string{"A: B", "X:Y"}
	headers, _ := headerStringsToPairs(args)
	assert.Equal([][]string{{"A", "B"}, []string{"X", "Y"}}, headers, "produces pairs")

	args = []string{"A-B", "X: Y"}
	headers, err = headerStringsToPairs(args)
	assert.Nil(headers)
	assert.NotNil(err)
}

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	// basic creation
	config, _ := NewConfig("http://localhost", []string{}, "{{.ID}}", false)
	assert.NotNil(config)

	// headers creation
	config, _ = NewConfig("http://localhost", []string{"X:Y", "A: B"}, "{{.ID}}", false)
	assert.NotNil(config)

	_, err := NewConfig("http://localhost", []string{"XY", "A: B"}, "{{.ID}}", false)
	assert.NotNil(err)

	// ill-formatted template
	_, err = NewConfig("http://localhost", []string{}, "{{.ID}", false)
	assert.NotNil(err)

	// bad accessor template
	_, err = NewConfig("http://localhost", []string{}, "{{.id}}", false)
	assert.NotNil(err)
}

func TestDefaultTemplateProducesJSON(t *testing.T) {
	assert := assert.New(t)
	config, _ := NewConfig("http://localhost", []string{}, DefaultTemplateString(), false)

	now := time.Unix(1000, 0)
	session := &Session{}
	session.Reset()
	session.Mail("to@you", smtp.MailOptions{})
	session.Rcpt("to@them")
	session.Rcpt("to2@them")
	session.data = "this is\nmy message"
	session.id = "fixed-id"
	session.receivedAt = now

	bytes, err := preparePostBody(session, config.template)
	assert.Nil(err)
	assert.Equal(
		`{"id":"fixed-id","received_at":"1970-01-01T10:16:40+10:00","from":"to@you","to":["to@them","to2@them"],"data":"this is\u000Amy message"}`,
		bytes.String(), "good json")

}
