package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// SubstItem is an item for substitution.
type SubstItem struct {
	Src  string `json:"src"`
	Repl string `json:"rep"`

	rxSrc *regexp.Regexp
	repl  []byte
}

func (si *SubstItem) replaceAll(data []byte) []byte {
	return si.rxSrc.ReplaceAll(data, si.repl)
}

func (m *SubstItem) prepare() error {
	rx, err := regexp.Compile(m.Src)
	if err != nil {
		return err
	}
	m.rxSrc = rx
	m.repl = []byte(m.Repl)
	return nil
}

// SubstItems is an set (array) of SubstItem.
type SubstItems []*SubstItem

func (items SubstItems) replaceAll(data []byte) []byte {
	for _, si := range items {
		data = si.replaceAll(data)
	}
	return data
}

func (items SubstItems) prepare() error {
	for _, m := range items {
		err := m.prepare()
		if err != nil {
			return err
		}
	}
	return nil
}

// SubstGroup is a group of SubstItem and target information.
type SubstGroup struct {
	MediaTypes []string   `json:"mediaTypes"`
	Path       string     `json:"path"`
	Items      SubstItems `json:"items"`

	mtypes map[string]struct{}
	rxPath *regexp.Regexp
}

var defaultMtypes = map[string]struct{}{
	"text/html": struct{}{},
}

func (sg *SubstGroup) isMatch(mt, path string) bool {
	if mt == "" {
		return false
	}
	if len(sg.Items) == 0 {
		return false
	}
	_, ok := sg.mtypes[mt]
	if !ok {
		return false
	}
	return sg.rxPath.MatchString(path)
}

func (sg *SubstGroup) replaceAll(data []byte) []byte {
	for _, items := range sg.Items {
		data = items.replaceAll(data)
	}
	return data
}

func (sg *SubstGroup) prepare() error {
	if len(sg.MediaTypes) == 0 {
		sg.mtypes = defaultMtypes
	} else {
		sg.mtypes = map[string]struct{}{}
		for _, typ := range sg.MediaTypes {
			sg.mtypes[typ] = struct{}{}
		}
	}

	rx, err := regexp.Compile(sg.Path)
	if err != nil {
		return err
	}
	sg.rxPath = rx

	for _, si := range sg.Items {
		err := si.prepare()
		if err != nil {
			return err
		}
	}

	return nil
}

// Substitutions is a set (array) of SubstGroup.
type Substitutions []*SubstGroup

func (s Substitutions) prepare() error {
	for _, sg := range s {
		err := sg.prepare()
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadSubstitutions loads a Substitutions from the file.
func LoadSubstitutions(name string) (Substitutions, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var v Substitutions
	err = json.NewDecoder(f).Decode(&v)
	if err != nil {
		return nil, err
	}

	err = v.prepare()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func mediaType(r *http.Response) string {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return ""
	}
	return mt
}

func readBody(r *http.Response) ([]byte, error) {
	return ioutil.ReadAll(r.Body)
}

func setBody(r *http.Response, b []byte) {
	br := bytes.NewReader(b)
	r.Body = ioutil.NopCloser(br)
	r.Header.Set("Content-Length", strconv.Itoa(br.Len()))
	r.ContentLength = int64(br.Len())
}

// Rewrite rewrites http.Response by substitutions.
func (s Substitutions) Rewrite(r *http.Response) error {
	mt := mediaType(r)
	path := r.Request.URL.Path
	var data []byte
	for _, sg := range s {
		if !sg.isMatch(mt, path) {
			continue
		}
		if data == nil {
			// read request body once.
			d, err := readBody(r)
			if err != nil {
				return err
			}
			data = d
		}
		data = sg.replaceAll(data)
	}
	if data == nil {
		return nil
	}
	setBody(r, data)
	return nil
}
