package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func currentUser() string {
	return user("--file", "fixtures/test.config")
}

func mockHomeEnv(dir string) {
	os.Setenv("HOME", dir)
}

func mockStdin(t *testing.T, path string) {
	fi, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdin = fi
}

func mockStdout(t *testing.T) (*os.File, *os.File) {
	oldstdout := os.Stdout
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error("Could not create temp file")
	}

	os.Stdout = tmp

	return tmp, oldstdout
}

func cleanupStdout(t *testing.T, tmp *os.File, stdout *os.File) {
	name := tmp.Name()
	err := tmp.Close()

	if err != nil {
		t.Error(err)
	}

	err = os.Remove(name)
	if err != nil {
		t.Error(err)
	}
	os.Stdout = stdout
}

func restorePearrc(t *testing.T, contents []byte) {
	p := path.Join(os.Getenv("HOME"), ".pearrc")
	err := ioutil.WriteFile(p, contents, os.ModeExclusive)
	if err != nil {
		t.Error(err)
	}
}

func TestPear(t *testing.T) {
	mockHomeEnv("fixtures/integration")
	mockStdin(t, "fixtures/integration/fullName.txt")
	tmp, oldstdout := mockStdout(t)
	originalPearrc := []byte("email: foo@example.com\ndevs:\n  dev1: Full Name A")
	restorePearrc(t, originalPearrc)
	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		restorePearrc(t, originalPearrc)
	}()

	os.Args = []string{"pear", "dev1", "dev2", "--file", "fixtures/test.config"}

	main()

	conf, err := readPearrc("fixtures/integration/.pearrc")
	if err != nil {
		t.Error(err)
	}

	if len(conf.Devs) != 2 {
		t.Error("Devs were not recorded")
	}

	expected := "Full Name A and Person B"
	if currentUser() != expected {
		t.Errorf("Expected %s got %s", expected, currentUser())
	}
}

func TestPearOneDevNoSavedEmail(t *testing.T) {
	mockHomeEnv("fixtures/integration")
	mockStdin(t, "fixtures/integration/email_prompt.txt")
	tmp, oldstdout := mockStdout(t)

	originalPearrc := []byte("devs:\n  dev1: Full Name A")
	restorePearrc(t, originalPearrc)
	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		restorePearrc(t, originalPearrc)
	}()

	os.Args = []string{"pear", "dev1", "--email", "foo@biz.net", "--file", "fixtures/test.config"}

	main()

	readConf, err := readPearrc("fixtures/integration/.pearrc")
	if err != nil {
		t.Fatal(err)
	}

	if readConf.Email != "dev@pear.biz" {
		t.Error("Email was not saved.")
	}
}

func TestCheckEmail(t *testing.T) {
	conf := Config{}

	mockStdin(t, "fixtures/integration/email_prompt.txt")
	tmp, oldstdout := mockStdout(t)

	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		originalPearrc := []byte("devs:\n  dev1: Full Name A")
		restorePearrc(t, originalPearrc)
	}()

	checkEmail(&conf)

	_, err := tmp.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Error(err)
	}

	output, err := ioutil.ReadAll(tmp)
	if err != nil {
		t.Error("Could not read from temp file")
	}

	if string(output) != "Please provide base author email:\n" {
		t.Errorf("Prompt was incorrect, got: %#v", string(output))
	}

	expected := "dev@pear.biz"
	if conf.Email != expected {
		t.Errorf("Expected %s, got %s", expected, conf.Email)
	}
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
	expectedFullName := "Person B"
	pair := []string{"knowndev", "newdev"}
	conf := &Config{
		Devs: map[string]string{
			"knowndev": "Known Dev",
		},
	}

	mockStdin(t, "fixtures/integration/fullName.txt")
	tmp, oldstdout := mockStdout(t)
	defer cleanupStdout(t, tmp, oldstdout)
	checkPair(pair, conf)

	_, err := tmp.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Error(err)
	}

	output, err := ioutil.ReadAll(tmp)
	if err != nil {
		t.Error("Could not read from temp file")
	}

	if string(output) != "Please enter a full name for newdev:\n" {
		t.Errorf("Question output was incorrect, got: %v", string(output))
	}

	fullName, ok := conf.Devs["newdev"]
	if !ok {
		t.Error("Dev was not found in conf")
	}

	if fullName != expectedFullName {
		t.Errorf("Expected %s got %s", expectedFullName, fullName)
	}
}
