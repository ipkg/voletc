package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

type Backend interface {
	// Get a key value map under a given prefix
	GetMap(string) (map[string][]byte, error)
	// Set key value map under the given prefix
	SetMap(string, map[string][]byte) error
	// Delete all keys under the given prefix
	DeleteMap(string) error
	// Determine if the key exists and has an assign value.
	IsKeyValid(string) bool
}

type ConfigKeys map[string][]byte

// Convert values to strings and return new map
func (ck ConfigKeys) ToString() map[string]string {
	nm := map[string]string{}
	for k, v := range ck {
		nm[k] = string(v)
	}
	return nm
}

type Template struct {
	Name string
	Body []byte
}

func NewTemplateFromKey(key string) *Template {
	pp := strings.Split(key, "/")
	if len(pp) > 1 {
		return &Template{Name: pp[1]}
	}

	return nil
}

type AppConfig struct {
	// Required fields
	Name    string
	Version string
	Env     string
	// Keys available to be applied to template
	Keys ConfigKeys
	// Available templates
	Templates []*Template
	// Backend consul, etcd ...
	be Backend
}

func NewAppConfigFromName(name string, be Backend) (*AppConfig, error) {
	a := &AppConfig{Templates: []*Template{}, Keys: ConfigKeys{}}

	pp := strings.Split(name, "-")
	if len(pp) < 3 {
		return nil, fmt.Errorf("invalid name format. must be <name>-<version>-<env>")
	}

	a.Env = pp[len(pp)-1]
	a.Version = pp[len(pp)-2]
	a.Name = strings.Join(pp[:len(pp)-2], "-")

	if a.Env == "" || a.Version == "" || a.Name == "" {
		return nil, fmt.Errorf("invalid name format. must be <name>-<version>-<env>")
	}

	var err error
	if be != nil {
		a.be = be
		err = a.fetchKeys()
	}

	return a, err
}

func (c *AppConfig) Meta() map[string]interface{} {
	return map[string]interface{}{
		"name":    c.Name,
		"version": c.Version,
		"env":     c.Env,
		"files":   len(c.Templates),
		"keys":    len(c.Keys),
	}
}

func (a *AppConfig) getOpaque(n string) string {
	return a.Name + "/" + a.Version + "/" + n
}

// Load data from backedn
func (a *AppConfig) fetchKeys() error {
	gm, err := a.be.GetMap(a.getOpaque(""))
	if err == nil {
		a.Set(gm)
	}

	return err
}

func (a *AppConfig) QualifiedName() string {
	return a.Name + "-" + a.Version + "-" + a.Env
}

// Create user passed keys to the backend
func (a *AppConfig) Init() error {
	m := a.buildBackendDataMap()
	// store to backend
	return a.be.SetMap(a.getOpaque(""), m)
}

// build payload from in mem data to write to backend
func (a *AppConfig) buildBackendDataMap() map[string][]byte {
	m := map[string][]byte{}
	// add appropriate key prefix
	for k, v := range a.Keys {
		m[a.Env+"/"+k] = v
	}
	// add template prefix
	for _, t := range a.Templates {
		m["templates/"+t.Name] = t.Body
	}
	return m
}

// Load data from backend, generate directory structure and
// config files under `basedir`
func (a *AppConfig) Load(basedir string) error {
	err := a.fetchKeys()
	if err == nil {

		keys := a.Keys.ToString()
		log.Println("Keys from store:", keys)

		for _, t := range a.Templates {
			d, err := parseTemplate(string(t.Body), keys)
			if err == nil {
				err = ioutil.WriteFile(basedir+"/"+t.Name, []byte(d), 0644)
			}

			if err != nil {
				return err
			}
		}
	}

	return err
}

// Destroy keys from the backend.
func (a *AppConfig) Destroy() error {
	return a.be.DeleteMap(a.getOpaque(a.Env + "/"))
}

// Set input data to  datastructure
func (a *AppConfig) Set(data map[string][]byte) {

	for key, v := range data {
		k := strings.TrimPrefix(key, a.getOpaque(""))

		switch {
		case strings.HasPrefix(k, "templates"):
			if t := NewTemplateFromKey(k); t != nil {
				t.Body = v
				a.Templates = append(a.Templates, t)
			}

		default:
			a.Keys[strings.TrimPrefix(k, a.Env+"/")] = v

		}
	}
}
