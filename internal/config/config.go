package config

import (
	"fmt"
	"regexp"
	"text/template"
)

// Config contains operatonal configuration values
type Config struct {
	Verbose  bool
	URL      string
	Headers  [][2]string
	Template *template.Template
}

// NewConfig creates an SMTP Pigeon configuration struct
func NewConfig(url string, headerArgs []string, templateString string, verbose bool) (*Config, error) {
	var headers [][2]string
	var tmpl *template.Template
	var err error

	headers, err = headerStringsToPairs(headerArgs)
	if err != nil {
		return nil, err
	}

	// check that the template will compile and will apply
	tmpl, err = template.New("post-template").Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse template: %v", err)
	}

	return &Config{
		Verbose:  verbose,
		URL:      url,
		Headers:  headers,
		Template: tmpl,
	}, nil
}

func headerStringsToPairs(headerArgs []string) ([][2]string, error) {
	var re = regexp.MustCompile(`(.+):\s*(.+)`)
	var headers [][2]string
	for _, arg := range headerArgs {
		match := re.FindStringSubmatch(arg)
		if len(match) == 0 {
			return nil, fmt.Errorf("Headers must be in the format `x: y`, got %q", arg)
		}
		pair := [2]string{match[1], match[2]}
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
