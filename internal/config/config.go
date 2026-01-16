package config

import (
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"os"
	"regexp"
	"text/template"
)

type HeaderPair struct {
	Key string
	Value *template.Template
}

// Config contains operatonal configuration values
type Config struct {
	Verbose  bool
	URL      *template.Template
	Headers  []HeaderPair
	Template *template.Template
}

// NewConfig creates an SMTP Pigeon configuration struct
func NewConfig(urlString string, headerArgs []string, templateString string, verbose bool) (*Config, error) {
	var headers []HeaderPair
	var urlTemplate *template.Template
	var bodyTemplate *template.Template
	var err error

	funcs := sprig.TxtFuncMap()
	funcs["env"] = os.Getenv

	headers, err = headerStringsToPairs(headerArgs)
	if err != nil {
		return nil, err
	}

	urlTemplate, err = template.New("url-template").Funcs(funcs).Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse url: %v", err)
	}

	bodyTemplate, err = template.New("post-template").Funcs(funcs).Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse template: %v", err)
	}

	return &Config{
		Verbose:  verbose,
		URL:      urlTemplate,
		Headers:  headers,
		Template: bodyTemplate,
	}, nil
}

func headerStringsToPairs(headerArgs []string) ([]HeaderPair, error) {
	var re = regexp.MustCompile(`(.+):\s*(.+)`)
	var headers []HeaderPair
	funcs := sprig.TxtFuncMap()
	funcs["env"] = os.Getenv
	for _, arg := range headerArgs {
		match := re.FindStringSubmatch(arg)
		if len(match) == 0 {
			return nil, fmt.Errorf("Headers must be in the format `x: y`, got %q", arg)
		}
		key := match[1]
		valueTemplate, err := template.New("header-template").Funcs(funcs).Parse(match[2])
		if err != nil {
			return nil, fmt.Errorf("Could not parse header: %v", err)
		}
		pair := HeaderPair{
			Key: key,
			Value: valueTemplate,
		}
		headers = append(headers, pair)
	}
	return headers, nil
}

// DefaultTemplateString returns the default JSON format template
func DefaultTemplateString() string {
	return `{"id":"{{.ID | js}}",` +
		`"timestamp":"{{.Timestamp.UTC.Format "2006-01-02T15:04:05Z07:00" | js }}",` +
		`"sender":"{{.Sender | js}}",` +
		`"recipients":[{{range $i, $e := .Recipients}}{{if $i}},{{end}}"{{$e | js}}"{{end}}],` +
		`"body":"{{.Body | js}}",` +
		`"subject":"{{.Header.Get "Subject"}}"}`
}
