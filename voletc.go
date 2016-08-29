package main

import (
	"fmt"
	"strings"
)

type VolEtc struct {
	be Backend
}

func (ve *VolEtc) Get(name string) (*AppConfig, error) {
	acfg, err := NewAppConfigFromName(name, ve.be)
	if err != nil {
		return nil, err
	}
	// doesn't really exist ???
	//if !acfg.HasMappedKeys() {
	if !acfg.Exists() {
		return nil, fmt.Errorf("not found: '%s'", name)
	}

	return acfg, nil
}

func (ve *VolEtc) List() (map[string]*AppConfig, error) {

	mp, err := ve.be.GetMap("")
	if err != nil {
		return nil, err
	}

	out := map[string]*AppConfig{}

	for k, _ := range mp {
		pp := strings.Split(k, "/")
		// TODO: Support app and version without environment
		// this would be needed to check for templates but no keys
		if pp[2] == "templates" {
			continue
		}
		name := fmt.Sprintf("%s-%s-%s", pp[0], pp[1], pp[2])

		acfg, err := NewAppConfigFromName(name, ve.be)
		if err != nil {
			return nil, err
		}
		// doesn't really exist ???
		//if !acfg.HasMappedKeys() {
		if !acfg.Exists() {
			continue
		}

		out[name] = acfg
	}

	return out, nil
}
