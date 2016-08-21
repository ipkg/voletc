package main

import (
	"strings"

	"github.com/hashicorp/consul/api"
)

type ConsulBackend struct {
	cfg    *api.Config
	client *api.Client

	prefix string
}

func NewConsulBackend(addr string, prefix string) (*ConsulBackend, error) {
	cbe := &ConsulBackend{
		cfg:    api.DefaultConfig(),
		prefix: prefix,
	}
	cbe.cfg.Address = addr

	var err error
	cbe.client, err = api.NewClient(cbe.cfg)

	return cbe, err
}

func (cb *ConsulBackend) getOpaque(key string) string {
	if cb.prefix == "" {
		return key
	}
	return cb.prefix + "/" + key
}

func (m *ConsulBackend) Set(key string, value []byte) error {
	kvc := m.client.KV()
	p := &api.KVPair{Key: m.getOpaque(key), Value: value}
	_, err := kvc.Put(p, nil)
	return err
}

func (m *ConsulBackend) SetMap(prefix string, mp map[string][]byte) error {
	kvc := m.client.KV()
	for k, v := range mp {
		p := &api.KVPair{Key: m.getOpaque(prefix + k), Value: v}
		if _, err := kvc.Put(p, nil); err != nil {
			return err
		}
	}

	return nil
}

func (m *ConsulBackend) GetMap(prefix string) (map[string][]byte, error) {
	kvc := m.client.KV()
	out := map[string][]byte{}

	kvs, _, err := kvc.List(m.getOpaque(prefix), nil)
	if err == nil {
		for _, kv := range kvs {
			key := strings.TrimPrefix(kv.Key, m.prefix+"/")
			out[key] = kv.Value
		}
	}
	return out, err
}

func (m *ConsulBackend) DeleteMap(prefix string) error {
	kvc := m.client.KV()
	_, err := kvc.DeleteTree(m.getOpaque(prefix), nil)
	return err
}
