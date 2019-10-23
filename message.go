package main

import (
	"encoding/json"
	"os"
	"regexp"
)

type Message struct {
	Src  string `json:"src"`
	Repl string `json:"rep"`

	rx   *regexp.Regexp
	repl []byte
}

func (m *Message) prepare() error {
	rx, err := regexp.Compile(m.Src)
	if err != nil {
		return err
	}
	m.rx = rx
	m.repl = []byte(m.Repl)
	return nil
}

type Messages []*Message

var allMsgs Messages

func readMessageFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&allMsgs)
	if err != nil {
		return err
	}
	for _, m := range allMsgs {
		err := m.prepare()
		if err != nil {
			return err
		}
	}
	return nil
}

func getMessages(path string) Messages {
	// TODO:
	return allMsgs
}
