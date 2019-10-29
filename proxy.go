package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Proxy is
type Proxy struct {
	proxy *httputil.ReverseProxy
	host  string
	subst Substitutions

	od func(*http.Request)
}

func newProxy(u *url.URL, subst Substitutions) *Proxy {
	rp := httputil.NewSingleHostReverseProxy(u)
	p := &Proxy{
		proxy: rp,
		host:  u.Host,
		subst: subst,
		od:    rp.Director,
	}
	rp.Director = p.filterRequest
	rp.ModifyResponse = p.filterResponse
	return p
}

func (p *Proxy) filterRequest(r *http.Request) {
	p.od(r)
	if p.host != "" {
		r.Host = p.host
	}
}

func (p *Proxy) filterResponse(r *http.Response) error {
	st := time.Now()
	err := p.subst.Rewrite(r)
	if err != nil {
		return err
	}
	d := time.Since(st)
	if optAccessLog {
		log.Printf("rewrite %s in %s", r.Request.URL.Path, d)
	}
	return nil
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.proxy.ServeHTTP(rw, req)
}
