package dispatch

import (
	"bytes"
	"fmt"
	"net/http"
	"net/mail"
	"text/template"
	"time"
	"github.com/rktjmp/smtp-pigeon/internal/config"
)

type TemplateData struct {
	ID         string
	Timestamp  time.Time
	Sender     string
	Recipients []string
	Data       string
	Header     mail.Header
	Body       string
}

type Endpoint struct {
	URL     *template.Template
	Headers []config.HeaderPair
}

func POST(endpoint *Endpoint, tmpl *template.Template, data *TemplateData) (int, error) {
	var bodyBuf bytes.Buffer
	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return 0, err
	}
	var urlBuf bytes.Buffer
	if err := endpoint.URL.Execute(&urlBuf, data); err != nil {
		return 0, err
	}
	var headers [][2]string
	for _, header := range endpoint.Headers {
		var valueBuf bytes.Buffer
		if err := header.Value.Execute(&valueBuf, data); err != nil {
			return 0, fmt.Errorf("could not execute header template: %q: %v", header.Key, err)
		}
		headers = append(headers, [2]string{
			header.Key,
			valueBuf.String(),
		})
	}
	resp, err := performPOSTRequest(urlBuf.String(), headers, &bodyBuf)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func performPOSTRequest(url string, headers [][2]string, body *bytes.Buffer) (*http.Response, error) {
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
