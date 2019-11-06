package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var proxyServer *httptest.Server

func TestMain(m *testing.M) {
	subst := Substitutions{
		{Path: `^/001/[^/]*\.html$`, Items: SubstItems{
			{SrcRx: `This is (\d+) YEN`, Repl: "これは $1 円です"},
		}},
	}
	err := subst.prepare()
	if err != nil {
		log.Fatalf("prepare failed: %s", err)
	}

	fs := httptest.NewServer(http.FileServer(http.Dir("./testdata")))
	u, err := url.Parse(fs.URL)
	if err != nil {
		fs.Close()
		log.Fatalf("failed to parse URL: %s", err)
	}

	p := newProxy(u, subst)

	proxyServer = httptest.NewServer(p)
	n := m.Run()
	proxyServer.Close()
	fs.Close()

	os.Exit(n)
}

func get(t *testing.T, c *http.Client, u string) []byte {
	t.Helper()
	resp, err := c.Get(u)
	if err != nil {
		t.Fatalf("failed to GET %s: %s", u, err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %s", err)
	}
	return b
}

func testGet(t *testing.T, path string, exp string) {
	t.Helper()
	u := proxyServer.URL + path
	act := get(t, proxyServer.Client(), u)
	if string(act) != exp {
		t.Fatalf("unexpected body\nexpected=%q\nactual=%q", exp, act)
	}
}

func Test_001_simple_substitutions(t *testing.T) {
	testGet(t, "/001/abc.html", "これは 1234 円です\n")
	testGet(t, "/001/def.html", "これは 999999 円です\n")
	testGet(t, "/001/ghi.html",
		"これは 1 円です\nこれは 2 円です\nこれは 3 円です\n")
}
