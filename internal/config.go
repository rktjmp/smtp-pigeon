package internal

import (
	"fmt"
	"regexp"
	"text/template"
)

// Config contains operatonal configuration values
type Config struct {
	verbose  bool
	url      string
	headers  [][]string
	template *template.Template
}

// NewConfig creates an SMTP Pigeon configuration struct
func NewConfig(url string, headerArgs []string, templateString string, verbose bool) (*Config, error) {
	var headers [][]string
	var tmpl *template.Template
	var err error

	headers, err = headerStringsToPairs(headerArgs)
	if err != nil {
		return nil, err
	}

	// check that the template will compile and will apply
	tmpl, err = template.New("JSONTemplate").Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse template: %v", err)
	}

	err = CheckTemplateExecute(tmpl)
	if err != nil {
		return nil, fmt.Errorf("Could not execute template: %v", err)
	}

	return &Config{
		verbose:  verbose,
		url:      url,
		headers:  headers,
		template: tmpl,
	}, nil
}

func headerStringsToPairs(headerArgs []string) ([][]string, error) {
	var re = regexp.MustCompile(`(.+):\s*(.+)`)
	var headers [][]string
	for _, arg := range headerArgs {
		match := re.FindStringSubmatch(arg)
		if len(match) == 0 {
			return nil, fmt.Errorf("Headers must be in the format `x: y`, got %q", arg)
		}
		pair := []string{match[1], match[2]}
		headers = append(headers, pair)
	}
	return headers, nil
}

// DefaultTemplateString returns the default JSON format template
func DefaultTemplateString() string {
	return `{"id":"{{.ID | js}}","received_at":"{{.ReceivedAt | js}}","from":"{{.From | js}}","to":[{{range $i, $e := .To}}{{if $i}},{{end}}"{{$e | js}}"{{end}}],"data":"{{.Data | js}}"}`
}
