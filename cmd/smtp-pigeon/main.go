package main

import (
	"flag"
	"fmt"
	"github.com/emersion/go-smtp"
	"github.com/rktjmp/smtp-pigeon/internal/backend"
	"github.com/rktjmp/smtp-pigeon/internal/config"
	"github.com/rktjmp/smtp-pigeon/internal/session"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

type stringSlice []string

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

type flags struct {
	help         bool // show help
	version      bool // show version
	prefixLogger bool // prefix logger with date-time

	mailDomain string // SMTP "hostname"

	listenHost string // listen server settings
	listenPort int

	endpointURL     string      // make post where
	endpointHeaders stringSlice // {header, header}

	templateString string // post what
}

func parseFlags() *flags {
	var defaultTemplate = config.DefaultTemplateString()
	flags := flags{}

	flag.BoolVar(&flags.version, "version", false, "Show version information")
	flag.BoolVar(&flags.help, "help", false, "View this text")
	flag.BoolVar(&flags.prefixLogger, "standalone-logging", false, "Prefix logs with date and time")

	flag.StringVar(&flags.mailDomain, "domain", "localhost", "Mail domain to reply to EHLO with")
	flag.StringVar(&flags.listenHost, "host", "127.0.0.1", "Address to bind to")
	flag.Var(&flags.endpointHeaders, "header", "Headers to attach to POST, must be in form \"Header: Value\", may be used multiple times")
	flag.IntVar(&flags.listenPort, "port", 1025, "Port to listen on")
	flag.StringVar(&flags.endpointURL, "url", "", "URL to make HTTP POST to, required")
	flag.StringVar(
		&flags.templateString,
		"template",
		defaultTemplate,
		`Template used to render POST body. Does not have to be JSON if you set the appropriate Content-Type header.
Can access:
  - ID (string),
  - RecievedAt (string),
  - From (string),
  - To (a list of strings) and
  - Data (multiline string).
`)

	flag.Parse()

	return &flags
}

func configureLog(prefix bool) {
	if prefix {
		log.SetPrefix("smtp-pigeon: ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	} else {
		log.SetPrefix("")
		log.SetFlags(0)
	}
}

func dryrun(cfg *config.Config) error {
	// run fake endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	// mostly use the real config, just point it at our fake server
	url := cfg.URL
	cfg.URL = server.URL

	log.SetOutput(io.Discard)

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

	log.SetOutput(os.Stderr)

	cfg.URL = url
	return err
}

func main() {
	flags := parseFlags()

	// some flags are exit or failure points
	switch {
	case flags.help:
		flag.PrintDefaults()
		os.Exit(0)
	case flags.version:
		fmt.Printf("smtp-pigeon version: %s (%s, %s, %s)\n", version, commit, builtBy, date)
		os.Exit(0)
	case flags.endpointURL == "":
		log.Println("Error: Must provide --url option")
		flag.PrintDefaults()
		os.Exit(1)
	}

	configureLog(flags.prefixLogger)
	config, err := config.NewConfig(
		flags.endpointURL,
		flags.endpointHeaders,
		flags.templateString,
		false,
	)
	if err != nil {
		log.Fatalln(err)
	}

	err = dryrun(config)
	if err != nil {
		log.Println("Dry run failed, refusing to start.")
		log.Fatalln(err)
	}

	be := backend.NewBackend(config)
	s := smtp.NewServer(be)
	s.Addr = fmt.Sprint(flags.listenHost, ":", flags.listenPort)
	s.Domain = flags.mailDomain
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("smtp-pigeon listening at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func (i *stringSlice) String() string {
	return "I'm just here so I don't get fined"
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}
