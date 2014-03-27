package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func currentUser() string {
	return user("--file", "fixtures/test.config")
}

func TestSetPairWithOneDev(t *testing.T) {
	setPair([]string{"user1"}, "--file", "fixtures/test.config")
	expected := "user1"

	if currentUser() != expected {
		t.Errorf("Expected %s got %s", expected, currentUser())
	}
}

func TestSetPairWithTwoDevs(t *testing.T) {
	pair := []string{"user1", "user2"}
	setPair(pair, "--file", "fixtures/test.config")
	expected := "user1 and user2"

	if currentUser() != expected {
		t.Errorf("Expected %s got %s", expected, currentUser())
	}
}

func TestReadPearrc(t *testing.T) {
	nonExistantPath := "fixtures/.fakepearrc"

	readPearrc(nonExistantPath)

	f, err := os.Open(nonExistantPath)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		f.Close()
		os.Remove(nonExistantPath)
	}()
}

func TestSavePearrc(t *testing.T) {
	expected := map[string]string{
		"dparker":   "Derek Parker",
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

func TestCheckPairWithUnknownDev(t *testing.T) {
	oldStdout := os.Stdout
	expectedFullName := "new developer"
	pair := []string{"knowndev", "newdev"}
	conf := &Config{
		Devs: map[string]string{
			"knowndev": "Known Dev",
		},
	}

	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error("Could not create temp file")
	}

	defer func() {
		name := tmp.Name()
		err := tmp.Close()

		if err != nil {
			t.Error(err)
		}

		err = os.Remove(name)
		if err != nil {
			t.Error(err)
		}
	}()

	os.Stdout = tmp

	fi, err := os.Open("fixtures/fullName.txt")
	if err != nil {
		t.Error("Could not open file")
	}

	defer fi.Close()

	os.Stdin = fi
	checkPair(pair, conf)
	os.Stdout = oldStdout

	_, err = tmp.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Error(err)
	}

	output, err := ioutil.ReadAll(tmp)
	if err != nil {
		t.Error("Could not read from temp file")
	}

	if string(output) != "Please enter your full name for newdev:\n" {
		t.Error("Question output was incorrect")
	}

	fullName, ok := conf.Devs["newdev"]
	if !ok {
		t.Error("Dev was not found in conf")
	}

	if fullName != expectedFullName {
		t.Errorf("Expected %s got %s", expectedFullName, fullName)
	}
}
