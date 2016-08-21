package main

import (
	"fmt"
)

func parseTemplate(contents string, m map[string]string) (string, error) {
	out := make([]byte, 0)

	sm := -1
	var i int
	for i = 0; i < len(contents); i++ {

		switch contents[i] {
		case '$':
			if contents[i-1] != '\\' && contents[i+1] == '{' {
				sm = i + 2
				i++
				continue
			}
		case '}':
			if contents[i-1] != '\\' && sm > -1 {

				key := contents[sm:i]
				out = append(out, []byte(m[key])...)
				sm = -1
				continue
			}

		}
		if sm < 0 {
			out = append(out, contents[i])
		}
	}

	err := validate(out)
	return string(out), err
}

func parseKeys(contents string) (map[string]bool, error) {
	err := validate([]byte(contents))
	if err != nil {
		return nil, err
	}

	tkeys := map[string]bool{}

	sm := -1
	var i int
	for i = 0; i < len(contents); i++ {

		switch contents[i] {
		case '$':
			if contents[i-1] != '\\' && contents[i+1] == '{' {
				sm = i + 2
				i++
				continue
			}
		case '}':
			if contents[i-1] != '\\' && sm > -1 {

				key := contents[sm:i]
				tkeys[key] = false
				sm = -1
				continue
			}

		}
	}

	return tkeys, nil
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
		return fmt.Errorf("curly brace mismatch")
	}
	return nil
}
