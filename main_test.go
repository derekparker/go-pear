package main

import (
	"testing"
)

func TestUser(t *testing.T) {
	if user("--file", "fixtures/test.config") != "test_user" {
		t.Errorf("Expected test_user got: %#v", globalUser())
	}
}

