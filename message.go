package main

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
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

func (msgs Messages) prepare() error {
	for _, m := range msgs {
		err := m.prepare()
		if err != nil {
			return err
		}
	}
	return nil
}

func readMessageFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	var msgs Messages
	err = json.NewDecoder(f).Decode(&msgs)
	if err != nil {
		return err
	}
	err = msgs.prepare()
	if err != nil {
		return err
	}
	defaultStructuredMessages[""] = msgs
	return nil
}

func getMessages(path string) Messages {
	return defaultStructuredMessages.getMessages(path)
}

type StructuredMessages map[string]Messages

func (sm StructuredMessages) load(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&sm)
	if err != nil {
		return err
	}
	err = sm.prepare()
	if err != nil {
		return err
	}
	return nil
}

func (sm StructuredMessages) prepare() error {
	for _, msgs := range sm {
		err := msgs.prepare()
		if err != nil {
			return err
		}
	}
	return nil
}

func (sm StructuredMessages) findKeys(s string) ([]string, int) {
	var keys []string
	var sum int
	for k, msgs := range sm {
		if strings.HasPrefix(s, k) {
			keys = append(keys, k)
			sum += len(msgs)
		}
	}
	sort.Strings(keys)
	return keys, sum
}

func (sm StructuredMessages) getMessages(path string) Messages {
	keys, sum := sm.findKeys(path)
	if sum == 0 {
		return nil
	}
	msgs := make(Messages, 0, sum)
	for _, k := range keys {
		msgs = append(msgs, sm[k]...)
	}
	return msgs
}

var defaultStructuredMessages = StructuredMessages{}
