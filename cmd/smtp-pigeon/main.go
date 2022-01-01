package main

import (
	"flag"
	"fmt"
	"github.com/emersion/go-smtp"
	"github.com/rktjmp/smtp-pigeon/internal"
	"log"
	"os"
	"time"
)

type stringSlice []string

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	var showHelp bool
	var showVersion bool
	var verbose bool
	var logWithPrefix bool
	var mailDomain string
	var listenHost string
	var listenPort int
	var endpointURL string
	var endpointHeaders stringSlice
	var templateString string
	var defaultTemplate = internal.DefaultTemplateString()

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showHelp, "help", false, "View this text")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&logWithPrefix, "standalone-logging", false, "Prefix logs with date and time")

	flag.StringVar(&mailDomain, "domain", "localhost", "Mail domain to reply to EHLO with")
	flag.StringVar(&listenHost, "host", "127.0.0.1", "Address to bind to")
	flag.Var(&endpointHeaders, "header", "Headers to attach to POST, must be in form \"Header: Value\", may be used multiple times")
	flag.IntVar(&listenPort, "port", 1025, "Port to listen on")
	flag.StringVar(&endpointURL, "url", "", "URL to make HTTP POST to, required")
	flag.StringVar(
		&templateString,
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

	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("smtp-pigeon version: %s (%s, %s, %s)\n", version, commit, builtBy, date)
		os.Exit(0)
	}

	if logWithPrefix {
		log.SetPrefix("smtp-pigeon: ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	} else {
		log.SetPrefix("")
		log.SetFlags(0)
	}

	if endpointURL == "" {
		log.Println("Error: Must provide --url option")
		flag.PrintDefaults()
		os.Exit(1)
	}

	config, err := internal.NewConfig(endpointURL, endpointHeaders, templateString, verbose)
	if err != nil {
		log.Fatalln(err)
	}

	be := internal.NewBackend(config)
	s := smtp.NewServer(be)
	s.Addr = fmt.Sprint(listenHost, ":", listenPort)
	s.Domain = mailDomain
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
