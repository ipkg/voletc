package main

import (
	"testing"
)

func Test_Exists(t *testing.T) {
	dcfg := NewDriverConfig(testConsulUri, "./testrun", "test-be-driver")

	be, err := NewBackend(dcfg)
	if err != nil {
		t.Fatal(err)
	}

	if be.KeyExists("voletc/bleep/bleep") {
		t.Fatal("key should not eixst")
	}
}
