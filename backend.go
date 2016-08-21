package main

import (
	"fmt"
)

type Backend interface {
	// Get a key value map under a given prefix
	GetMap(string) (map[string][]byte, error)
	// Set key value map under the given prefix
	SetMap(string, map[string][]byte) error
	// Delete all keys under the given prefix
	DeleteMap(string) error
}

func NewBackend(dcfg *DriverConfig) (Backend, error) {
	var (
		be  Backend
		err error
	)

	switch dcfg.BackendType {
	case "consul":
		be, err = NewConsulBackend(dcfg.BackendAddr, dcfg.Prefix)

	default:
		err = fmt.Errorf("backend not supported: %s", dcfg.BackendType)

	}

	return be, err
}
