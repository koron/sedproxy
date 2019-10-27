package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	optTarget    string
	optAddr      string
	optHost      string
	optAccessLog bool
	optMsgs      string
)

var defaultSubstitution Substitutions

func run(ctx context.Context) error {
	flag.StringVar(&optTarget, "target",
		os.Getenv("REVERSE_PROXY_TARGET_URL"),
		`reverse proxy target URL`)
	flag.StringVar(&optAddr, "addr", ":8000",
		`reverse proxy server address and port`)
	flag.BoolVar(&optAccessLog, "accesslog", false, `output access log`)
	flag.StringVar(&optMsgs, "messages", "", `message file`)
	flag.Parse()

	if optTarget == "" {
		return errors.New("no targets. check -target or REVERSE_PROXY_TARGET_URL env")
	}
	tu, err := url.Parse(optTarget)
	if err != nil {
		return err
	}

	if optMsgs == "" {
		return errors.New("no messages, check -messages")
	}
	defaultSubstitution, err = LoadSubstitutions(optMsgs)

	p := newProxy(tu, defaultSubstitution)
	srv := &http.Server{
		Addr:    optAddr,
		Handler: p,
	}
	log.Printf("reveser proxy is listening %s\n", optAddr)
	return srv.ListenAndServe()
}
