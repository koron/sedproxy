package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

type contentEncode interface {
	decode(io.Reader) ([]byte, error)
	encode([]byte) (io.ReadCloser, int, error)
}

type identityEncoder struct{}

func (ie *identityEncoder) decode(r io.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	return b, err
}

func (ie *identityEncoder) encode(b []byte) (io.ReadCloser, int, error) {
	return ioutil.NopCloser(bytes.NewReader(b)), len(b), nil
}

type gzipEncode struct{}

func (ge *gzipEncode) decode(r io.Reader) ([]byte, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	b, err := ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (ge *gzipEncode) encode(b []byte) (io.ReadCloser, int, error) {
	bb := &bytes.Buffer{}
	gw := gzip.NewWriter(bb)
	_, err := gw.Write(b)
	if err != nil {
		return nil, 0, err
	}
	err = gw.Close()
	if err != nil {
		return nil, 0, err
	}
	return ioutil.NopCloser(bb), bb.Len(), nil
}

type deflateEncode struct{}

func (de *deflateEncode) decode(r io.Reader) ([]byte, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	b, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (de *deflateEncode) encode(b []byte) (io.ReadCloser, int, error) {
	bb := &bytes.Buffer{}
	zw := zlib.NewWriter(bb)
	_, err := zw.Write(b)
	if err != nil {
		return nil, 0, err
	}
	err = zw.Close()
	if err != nil {
		return nil, 0, err
	}
	return ioutil.NopCloser(bb), bb.Len(), nil
}

type brotliEncode struct{}

func (be *brotliEncode) decode(r io.Reader) ([]byte, error) {
	br := brotli.NewReader(r)
	b, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (be *brotliEncode) encode(b []byte) (io.ReadCloser, int, error) {
	bb := &bytes.Buffer{}
	gw := brotli.NewWriter(bb)
	_, err := gw.Write(b)
	if err != nil {
		return nil, 0, err
	}
	err = gw.Close()
	if err != nil {
		return nil, 0, err
	}
	return ioutil.NopCloser(bb), bb.Len(), nil
}

type response struct {
	res *http.Response
	ce  contentEncode
}

func newResponse(res *http.Response) *response {
	// check Content-Encoding, to support gzip and deflate.
	var ce contentEncode
	switch strings.ToLower(res.Header.Get("Content-Encoding")) {
	case "gzip":
		ce = &gzipEncode{}
	case "deflate":
		ce = &deflateEncode{}
	case "br":
		ce = &brotliEncode{}
	default:
		ce = &identityEncoder{}
	}

	return &response{
		res: res,
		ce:  ce,
	}
}

func (r *response) mediaType() string {
	ct := r.res.Header.Get("Content-Type")
	if ct == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return ""
	}
	return mt
}

func (r *response) path() string {
	return r.res.Request.URL.Path
}

func (r *response) readBody() ([]byte, error) {
	return r.ce.decode(r.res.Body)
}

func (r *response) setBody(b []byte) error {
	body, clen, err := r.ce.encode(b)
	if err != nil {
		return err
	}
	r.res.Body = body
	r.res.Header.Set("Content-Length", strconv.Itoa(clen))
	r.res.ContentLength = int64(clen)
	return nil
}
