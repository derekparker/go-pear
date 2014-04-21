package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/libgit2/git2go"
)

func TestPear(t *testing.T) {
	mockHomeEnv("fixtures/integration")
	tmpstdin := mockStdin(t, "Person B")
	tmp, oldstdout := mockStdout(t)
	pearrc := createPearrc(t, []byte("email: foo@example.com\ndevs:\n  deva: Full Name A"))
	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		closeFile(tmpstdin)
		closeFile(pearrc)
	}()

	os.Args = []string{"pear", "DevB", "DevA", "--file", "fixtures/test.config"}

	main()

	conf, err := readPearrc("fixtures/integration/.pearrc")
	if err != nil {
		t.Error(err)
	}

	if len(conf.Devs) != 2 {
		t.Error("Devs were not recorded")
	}

	expectedUser := "Full Name A and Person B"
	gitconfig := initTestGitConfig("fixtures/test.config", t)
	actualUser := username(gitconfig)
	if actualUser != expectedUser {
		t.Errorf("Expected %s got %s", expectedUser, actualUser)
	}

	expectedEmail := "foo+deva+devb@example.com"
	actualEmail := email(gitconfig)
	if actualEmail != expectedEmail {
		t.Errorf("Expected %s got %s", expectedEmail, actualEmail)
	}
}

func TestPearOneDevNoSavedEmail(t *testing.T) {
	mockHomeEnv("fixtures/integration")
	tmpstdin := mockStdin(t, "dev@pear.biz")
	tmp, oldstdout := mockStdout(t)

	pearrc := createPearrc(t, []byte("devs:\n  dev1: Full Name A"))
	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		closeFile(tmpstdin)
		closeFile(pearrc)
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

func TestPearWithinSubdirectory(t *testing.T) {
	mockHomeEnv("fixtures/integration")
	pearrc := createPearrc(t, []byte("email: foo@example.com\ndevs:\n  deva: Full Name A\n  devb: Full Name B"))
	defer closeFile(pearrc)

	withinStubRepo(t, "foo", func(conf *git.Config) {
		err := os.MkdirAll("bar", os.ModePerm|os.ModeExclusive|os.ModeDir)
		if err != nil {
			t.Fatal(err)
		}

		err = os.Chdir("bar")
		if err != nil {
			t.Fatal(err)
		}

		os.Args = []string{"pear", "DevB", "DevA"}
		main()

		conf.Refresh()

		expected := "Full Name A and Full Name B"
		if usr := username(conf); usr != expected {
			t.Errorf("Expected %s, got %s", expected, usr)
		}
	})
}

func TestCheckEmail(t *testing.T) {
	conf := Config{}

	mockHomeEnv("fixtures/integration")
	tempstdin := mockStdin(t, "dev@pear.biz")
	tmp, oldstdout := mockStdout(t)
	pearrc := createPearrc(t, []byte("devs:\n  dev1: Full Name A"))

	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		closeFile(tempstdin)
		closeFile(pearrc)
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
	gitconfig := initTestGitConfig("fixtures/test.config", t)

	setPair("foo@example.com", []string{"user1"}, gitconfig)
	expected := "user1"
	actual := username(gitconfig)

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

func TestSetPairWithTwoDevs(t *testing.T) {
	pair := []string{"user1", "user2"}
	formattedEmail := formatEmail("dev@example.com", pair)
	gitconfig := initTestGitConfig("fixtures/test.config", t)

	setPair(formattedEmail, pair, gitconfig)
	expectedUser := "user1 and user2"
	actualUser := username(gitconfig)
	expectedEmail := "dev+user1+user2@example.com"
	actualEmail := email(gitconfig)

	if actualUser != expectedUser {
		t.Errorf("Expected %s got %s", expectedUser, actualUser)
	}

	if actualEmail != expectedEmail {
		t.Errorf("Expected %s got %s", expectedEmail, actualEmail)
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

	tmpstdin := mockStdin(t, "Person B")
	tmp, oldstdout := mockStdout(t)
	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		closeFile(tmpstdin)
	}()
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

func TestEmailFormat(t *testing.T) {
	tests := []struct {
		email    string
		devs     []string
		expected string
	}{
		{"dev@example.com", []string{"dev1"}, "dev+dev1@example.com"},
		{"dev@example.com", []string{"dev1", "dev2"}, "dev+dev1+dev2@example.com"},
		{"dev@example.com", []string{"dev1", "dev2", "dev3"}, "dev+dev1+dev2+dev3@example.com"},
	}

	for _, test := range tests {
		actual := formatEmail(test.email, test.devs)

		if actual != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, actual)
		}
	}
}
