package main

import (
	"testing"
)

func currentUser() string {
	return user("--file", "fixtures/test.config")
}

func TestSetPairWithOneDev(t *testing.T) {
	setPair("user1", "", "--file", "fixtures/test.config")
	expected := "user1"

	if currentUser() != expected {
		t.Errorf("Expected %s got %s", expected, currentUser())
	}
}

func TestSetPairWithTwoDevs(t *testing.T) {
	setPair("user1", "user2", "--file", "fixtures/test.config")
	expected := "user1 and user2"

	if currentUser() != expected {
		t.Errorf("Expected %s got %s", expected, currentUser())
	}
}
