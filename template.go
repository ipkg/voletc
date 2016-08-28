package main

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

type Template struct {
	Name string `json:"name"`
	Body []byte `json:"body"`
	Sha1 string `json:"sha1"`

	rendered []byte
}

func NewTemplateFromKey(key string) *Template {
	pp := strings.Split(key, "/")
	if len(pp) > 1 {
		return &Template{Name: pp[1]}
	}

	return nil
}

func (t *Template) SetBody(b []byte) {
	t.Body = b
	t.Sha1 = fmt.Sprintf("%x", sha1.Sum(t.Body))
	t.rendered = b
	return
}

func (t *Template) Render(m map[string]string) ([]byte, error) {
	out := make([]byte, 0)

	sm := -1
	var i int
	for i = 0; i < len(t.Body); i++ {

		switch t.Body[i] {
		case '$':
			if t.Body[i-1] != '\\' && t.Body[i+1] == '{' {
				sm = i + 2
				i++
				continue
			}
		case '}':
			if t.Body[i-1] != '\\' && sm > -1 {

				key := string(t.Body[sm:i])
				out = append(out, []byte(m[key])...)
				sm = -1
				continue
			}

		}
		if sm < 0 {
			out = append(out, t.Body[i])
		}
	}

	err := validate(out)
	if err == nil {
		t.rendered = out
	}

	return out, err
}

func (t *Template) clearRendered() {
	t.rendered = t.Body
}

// Extract keys from template
func (t *Template) Keys() (map[string]bool, error) {
	//func parseKeys(contents string) (map[string]bool, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}

	tkeys := map[string]bool{}

	sm := -1
	var i int
	for i = 0; i < len(t.Body); i++ {

		switch t.Body[i] {
		case '$':
			if t.Body[i-1] != '\\' && t.Body[i+1] == '{' {
				sm = i + 2
				i++
				continue
			}
		case '}':
			if t.Body[i-1] != '\\' && sm > -1 {
				tkeys[string(t.Body[sm:i])] = false
				sm = -1
				continue
			}

		}
	}

	return tkeys, nil
}

// Validate curly braces
func (t *Template) Validate() error {
	return validate(t.Body)
}

func validate(in []byte) error {
	st := []byte{}
	for _, v := range in {
		switch v {
		case '{':
			st = append(st, v)
		case '}':
			st = st[:len(st)-1]
		}
	}

	if len(st) != 0 {
		return fmt.Errorf("missing end brace: '%s'", st)
	}
	return nil
}
