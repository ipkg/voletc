package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

var (
	errInvalidConfName = fmt.Errorf("invalid name: <name>-<version>-<env>")
)

type ConfigKeys map[string][]byte

// Convert values to strings and return new map
func (ck ConfigKeys) ToString() map[string]string {
	nm := map[string]string{}
	for k, v := range ck {
		nm[k] = string(v)
	}
	return nm
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
	var err error

	if a.Name, a.Version, a.Env, err = parseAppName(name); err == nil {
		if be != nil {
			a.be = be
			err = a.Load()
		}
	}
	return a, err
}

func (ac *AppConfig) AddTemplate(t *Template) error {
	found := false
	for _, v := range ac.Templates {
		if v.Name == t.Name && v.Sha1 == t.Sha1 {
			found = true
		}
	}
	if !found {

		keys, err := t.Keys()
		if err != nil {
			return err
		}

		ac.Templates = append(ac.Templates, t)
		// add template keys
		for k, _ := range keys {
			if _, ok := ac.Keys[k]; !ok {
				ac.Keys[k] = nil
			}
		}
	}

	return nil
}

func (c *AppConfig) Exists() bool {
	return c.be.KeyExists(c.getOpaque(c.Env))
}

func (c *AppConfig) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"id":      c.QualifiedName(),
		"name":    c.Name,
		"version": c.Version,
		"env":     c.Env,
		"files":   len(c.Templates),
		"keys":    len(c.Keys),
	}
}

// Load data from backedn
func (a *AppConfig) Load() error {
	gm, err := a.be.GetMap(a.getOpaque(""))
	if err == nil {
		a.Set(gm)
	}

	return err
}

func (a *AppConfig) QualifiedName() string {
	return a.Name + "-" + a.Version + "-" + a.Env
}

// Store in mem datastructure to backend
func (a *AppConfig) Commit() error {
	m := a.buildBackendDataMap()
	// store to backend
	return a.be.SetMap(a.getOpaque(""), m)
}

// Load data from backend, generate directory structure and
// rendered config files under `basedir`
func (a *AppConfig) Generate(basedir string) error {
	err := a.Load()
	if err == nil {
		keys := a.Keys.ToString()
		for _, t := range a.Templates {
			rendered, err := t.Render(keys)
			if err == nil {
				if err = ioutil.WriteFile(basedir+"/"+t.Name, rendered, 0644); err == nil {
					continue
				}
			}

			break
		}
	}

	return err
}

func (a *AppConfig) cacheRender() {
	keys := a.Keys.ToString()
	for _, t := range a.Templates {
		if _, err := t.Render(keys); err != nil {
			log.Println("ERR", err)
		}
	}
}

// Destroy keys from the backend.
func (a *AppConfig) Destroy() error {
	return a.be.DeleteMap(a.getOpaque(a.Env + "/"))
}

// Set input data to  datastructure.  Strip key prefixes before setting
func (a *AppConfig) Set(data map[string][]byte) error {

	for key, v := range data {
		k := strings.TrimPrefix(key, a.getOpaque(""))
		switch {

		case strings.HasPrefix(k, "templates"):
			if t := NewTemplateFromKey(k); t != nil {
				t.SetBody(v)
				a.AddTemplate(t)
			}

		default:
			a.Keys[strings.TrimPrefix(k, a.Env+"/")] = v

		}
	}

	return nil
}

func (a *AppConfig) getOpaque(n string) string {
	return a.Name + "/" + a.Version + "/" + n
}

// build payload from in mem data to write to backend
// it adds the prefix to each key and returns a new map
func (a *AppConfig) buildBackendDataMap() map[string][]byte {
	m := map[string][]byte{}
	// Add environment prefix to each config key
	for k, v := range a.Keys {
		m[a.Env+"/"+k] = v
	}
	// Add prefix to template keys
	for _, t := range a.Templates {
		m["templates/"+t.Name] = t.Body
	}
	return m
}

func parseAppName(n string) (name, version, env string, err error) {
	pp := strings.Split(n, "-")
	if len(pp) < 3 {
		err = errInvalidConfName
		return
	}

	env = pp[len(pp)-1]
	version = pp[len(pp)-2]
	name = strings.Join(pp[:len(pp)-2], "-")

	if env == "" || version == "" || name == "" {
		err = errInvalidConfName
	}
	return
}
