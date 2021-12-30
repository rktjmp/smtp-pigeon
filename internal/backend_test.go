package internal

import (
	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogin(t *testing.T) {
	assert := assert.New(t)

	be := NewBackend(&Config{})
	session, err := be.Login(&smtp.ConnectionState{}, "anything", "accepted")
	assert.Nil(err)
	assert.NotNil(session)
}

func TestAnonymousLogin(t *testing.T) {
	assert := assert.New(t)

	be := NewBackend(&Config{})
	session, err := be.AnonymousLogin(&smtp.ConnectionState{})
	assert.Nil(err)
	assert.NotNil(session)
}
