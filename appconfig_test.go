package main

import (
	"testing"
)

func Test_parseAppName_Error(t *testing.T) {
	_, _, _, err := parseAppName("foo--")
	if err == nil {
		t.Fatal("should fail")
	}

	if _, _, _, err := parseAppName("foo"); err == nil {
		t.Fatal("should fail")
	}

	if _, _, _, err := parseAppName("foo-"); err == nil {
		t.Fatal("should fail")
	}

	if _, _, _, err := parseAppName(""); err == nil {
		t.Fatal("should fail")
	}

	if _, _, _, err := parseAppName("foo-bar"); err == nil {
		t.Fatal("should fail")
	}
}
