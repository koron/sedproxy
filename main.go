package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	optHost      string
	optAccessLog bool
	optMsgs      string
	optMsgs2     string
)

func run(ctx context.Context) error {
	target := flag.String("target",
		os.Getenv("REVERSE_PROXY_TARGET_URL"),
		`reverse proxy target URL`)
	addr := flag.String("addr", ":8000",
		`reverse proxy server address and port`)
	flag.BoolVar(&optAccessLog, "accesslog", false, `output access log`)
	flag.StringVar(&optMsgs, "messages", "", `message file, exclusive with -structuredmessages"`)
	flag.StringVar(&optMsgs2, "structuredmessages", "", `structured message file, exclusive with -messages`)
	flag.Parse()

	if *target == "" {
		return errors.New("no targets. check -target or REVERSE_PROXY_TARGET_URL env")
	}
	tu, err := url.Parse(*target)
	if err != nil {
		return err
	}
	optHost = tu.Host

	if optMsgs == "" && optMsgs2 == "" {
		return errors.New("no messages. check -messages or -structuredmessages")
	}
	if optMsgs != "" && optMsgs2 != "" {
		return errors.New("choose one of -messages or -structuredmessages")
	}
	if optMsgs != "" {
		err := readMessageFile(optMsgs)
		if err != nil {
			return err
		}
	}
	if optMsgs2 != "" {
		err := defaultStructuredMessages.load(optMsgs2)
		if err != nil {
			return err
		}
	}

	rp := httputil.NewSingleHostReverseProxy(tu)
	rp.Director = wrapDirector(rp.Director)
	rp.ModifyResponse = filterResponse

	srv := &http.Server{
		Addr:    *addr,
		Handler: rp,
	}

	log.Printf("reveser proxy is listening %s\n", *addr)
	return srv.ListenAndServe()
}

const mtHTML = "text/html"

func isHTML(s string) bool {
	if s == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(s)
	if err != nil {
		log.Printf("failed to parse media type: %s", err)
		return false
	}
	return mt == mtHTML
}

func wrapDirector(orig func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		orig(r)
		if optHost != "" {
			r.Host = optHost
		}
	}
}

func filterResponse(r *http.Response) error {
	ct := r.Header.Get("Content-Type")
	if !isHTML(ct) {
		return nil
	}
	st := time.Now()
	err := modifyResponse(r)
	if err != nil {
		return err
	}
	d := time.Since(st)
	if optAccessLog {
		log.Printf("rewrite %s in %s", r.Request.URL.Path, d)
	}
	return nil
}

func modifyResponse(r *http.Response) error {
	msgs := getMessages(r.Request.URL.Path)
	b, err := replaceBody(r.Body, msgs)
	if err != nil {
		return err
	}
	br := bytes.NewReader(b)
	r.Body = ioutil.NopCloser(br)
	r.Header.Set("Content-Length", strconv.Itoa(br.Len()))
	r.ContentLength = int64(br.Len())
	return nil
}

func replaceBody(src io.ReadCloser, msgs Messages) ([]byte, error) {
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}
	for _, m := range msgs {
		b = m.rx.ReplaceAll(b, m.repl)
	}
	return b, nil
}
