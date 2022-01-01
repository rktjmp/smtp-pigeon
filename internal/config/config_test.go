package config_test

import (
	"encoding/json"
	"github.com/emersion/go-smtp"
	"github.com/rktjmp/smtp-pigeon/internal/config"
	"github.com/rktjmp/smtp-pigeon/internal/session"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	// basic creation
	cfg, _ := config.NewConfig("http://localhost", []string{}, "{{.ID}}", false)
	assert.NotNil(cfg)

	// headers creation
	cfg, _ = config.NewConfig("http://localhost", []string{"X:Y", "A: B"}, "{{.ID}}", false)
	assert.NotNil(cfg)

	_, err := config.NewConfig("http://localhost", []string{"XY", "A: B"}, "{{.ID}}", false)
	assert.NotNil(err)

	// ill-formatted template
	_, err = config.NewConfig("http://localhost", []string{}, "{{.ID}", false)
	assert.NotNil(err)
}

func TestDefaultTemplateProducesJSON(t *testing.T) {
	assert := assert.New(t)

	// run fake endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// body as expected
		var result map[string]interface{}
		bytes, err := io.ReadAll(r.Body)
		assert.Nil(err)
		err = json.Unmarshal(bytes, &result)
		assert.Nil(err)
		log.Println(string(bytes))
		log.Println(result)
		assert.NotEqual("", result["id"])
		assert.NotEqual("", result["timestamp"])
		assert.Equal("freeman@mailhub.bm.net", result["sender"])
		assert.Equal([]interface{}([]interface{}{"vance@mailhub.bm.net", "kleiner@mailhub.bm.net"}), result["recipients"])
		assert.Equal("ON MY WAY", result["subject"])
		assert.Equal("hey guys running L8 2DAY\non the tram now", result["body"])
	}))
	defer server.Close()
	cfg, _ := config.NewConfig(server.URL, []string{}, config.DefaultTemplateString(), false)

	data := `Subject: ON MY WAY
From: Gordon Freeman <freeman@materials.blackmesa.com>
To: Eli Vance <vance@materials.blackmesa.com>
Cc: Issac Kleiner <kleiner@materials.blackmesa.com>

hey guys running L8 2DAY
on the tram now`

	session := session.NewSession(cfg)
	session.Mail("freeman@mailhub.bm.net", smtp.MailOptions{})
	session.Rcpt("vance@mailhub.bm.net")
	session.Rcpt("kleiner@mailhub.bm.net")
	err := session.Data(strings.NewReader(data))
	assert.Nil(err)
}
