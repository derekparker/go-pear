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

func TestSavePearrc(t *testing.T) {
	expected := map[string]string{
		"dparker": "Derek Parker",
		"chriserin": "Chris Erin",
	}

	conf := Config{
		Devs: expected,
	}

	err := savePearrc(&conf, "fixtures/.pearrc")
	if err != nil {
		t.Fatal(err)
	}

	readConf, err := readPearrc("fixtures/.pearrc")
	if err != nil {
		t.Fatal(err)
	}

	actual := readConf.Devs
	if len(actual) != len(expected) {
		t.Error("Did not read devs")
	}

	for username, dev := range expected {
		if actual[username] != dev {
			t.Errorf("Expected %s got %s", dev, actual[username])
		}
	}
}
