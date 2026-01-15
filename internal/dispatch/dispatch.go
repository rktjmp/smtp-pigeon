package dispatch

import (
	"bytes"
	"fmt"
	"net/http"
	"net/mail"
	"text/template"
	"time"
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
	Headers [][2]string
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
	resp, err := performPOSTRequest(urlBuf.String(), endpoint.Headers, &bodyBuf)
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
