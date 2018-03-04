package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"
)

func TestPearTwoDevsOneWithoutEmail(t *testing.T) {
	mockHomeEnv("fixtures/integration")
	tmpstdin := mockStdinUser(t, "Person B", "personb@example.com")
	tmp, oldstdout := mockStdout(t)
	pearrc := createPearrc(t, []byte("email: foo@example.com\ndevs:\n  deva:\n    name: Full Name A"))

	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		closeFile(tmpstdin)
		closeFile(pearrc)
	}()

	withinStubRepo(t, "foo", func() {
		os.Args = []string{"pear", "DevB", "DevA"}

		main()

		conf, err := readPearrc(path.Join(os.Getenv("HOME"), ".pearrc"))
		if err != nil {
			t.Error(err)
		}

		if len(conf.Devs) != 2 {
			t.Error("Devs were not recorded")
		}

		expectedUser := "Full Name A and Person B"

		actualUser := username()

		if actualUser != expectedUser {
			t.Errorf("Expected %s got %s", expectedUser, actualUser)
		}

		expectedEmail := "foo+deva+devb@example.com"
		actualEmail := email()
		if actualEmail != expectedEmail {
			t.Errorf("Expected %s got %s", expectedEmail, actualEmail)
		}
	})
}

func TestPearWithinSubdirectory(t *testing.T) {
	pearrc := createPearrc(t, []byte("email: foo@example.com\ndevs:\n  deva:\n    name: Full Name A\n    email: a@a.com\n  devb:\n    name: Full Name B\n    email: b@b.com"))
	defer closeFile(pearrc)

	withinStubRepo(t, "foo", func() {
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

		expected := "Full Name A and Full Name B"
		if usr := username(); usr != expected {
			t.Errorf("Expected %s, got %s", expected, usr)
		}
	})
}

func TestCheckEmail(t *testing.T) {
	conf := Config{}

	mockHomeEnv("fixtures/integration")
	tempstdin := mockStdinEmail(t, "dev@pear.biz")
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
	withinStubRepo(t, "foo", func() {
		setPair("foo@example.com", []Dev{Dev{Name: "user1", Email: "email1"}}, []string{"user1"})
		expected := "user1"
		actual := username()

		if actual != expected {
			t.Errorf("Expected %s got %s", expected, actual)
		}
	})
}

func TestSetPairWithTwoDevs(t *testing.T) {
	withinStubRepo(t, "foo", func() {
		pair := []string{"user1", "user2"}

		devValues := []Dev{
			Dev{Name: "user1"},
			Dev{Name: "user2"},
		}

		formattedEmail := formatEmail("dev@example.com", pair)

		setPair(formattedEmail, devValues, []string{"user1", "user2"})
		expectedUser := "user1 and user2"
		actualUser := username()
		expectedEmail := "dev+user1+user2@example.com"
		actualEmail := email()

		if actualUser != expectedUser {
			t.Errorf("Expected %s got %s", expectedUser, actualUser)
		}

		if actualEmail != expectedEmail {
			t.Errorf("Expected %s got %s", expectedEmail, actualEmail)
		}
	})
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
	expected := map[string]Dev{
		"dparker":   Dev{Name: "Derek Parker"},
		"chriserin": Dev{Name: "Chris Erin"},
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
		Devs: map[string]Dev{
			"knowndev": Dev{Name: "Known Dev", Email: "knowndev@example.com"},
		},
	}

	tmpstdin := mockStdinUser(t, "Person B", "personb@example.com")
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

	fullName, ok := conf.Devs["newdev"]
	if !ok {
		t.Error("Dev was not found in conf")
	}

	if fullName.Name != expectedFullName {
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

func TestAppendCommitMessageAugmentation(t *testing.T) {
	withinStubRepo(t, "foo", func() {
		err := ioutil.WriteFile(".git/config", []byte("[pear]\n\tdevs = mattpolito\n"), os.ModeExclusive)
		var revision string
		messageSource := "commit"

		commitMessageFile, _ := ioutil.TempFile("", "")
		commitMessageFile.WriteString("abc123")
		commitMessageFile.Close()

		defer os.Remove(commitMessageFile.Name())

		devs := map[string]Dev{
			"mattpolito": Dev{Name: "Matt Polito", Email: "matt.polito@gmail.com"},
		}

		conf := &Config{
			Devs: devs,
		}

		augmentCommitMessage(
			commitMessageFile.Name(),
			messageSource,
			revision,
			conf,
		)

		contents, err := ioutil.ReadFile(commitMessageFile.Name())
		if err != nil {
			log.Fatal(err)
		}

		if !strings.Contains(string(contents), "Co-authored-by: Matt Polito <matt.polito@gmail.com>") {
			t.Error("The Co-author line was not included in the commit message")
		}

		if !strings.Contains(string(contents), "abc123") {
			t.Error("The existing contents were replaced")
		}
	})
}

func TestAppendCommitMessageAugmentationWithMerge(t *testing.T) {
	withinStubRepo(t, "foo", func() {
		err := ioutil.WriteFile(".git/config", []byte("[pear]\n\tdevs = mattpolito\n"), os.ModeExclusive)
		var revision string
		messageSource := "merge"

		commitMessageFile, _ := ioutil.TempFile("", "")
		commitMessageFile.WriteString("abc123")
		commitMessageFile.Close()

		defer os.Remove(commitMessageFile.Name())

		devs := map[string]Dev{
			"mattpolito": Dev{Name: "Matt Polito", Email: "matt.polito@gmail.com"},
		}

		conf := &Config{
			Devs: devs,
		}

		augmentCommitMessage(
			commitMessageFile.Name(),
			messageSource,
			revision,
			conf,
		)

		contents, err := ioutil.ReadFile(commitMessageFile.Name())
		if err != nil {
			log.Fatal(err)
		}

		if strings.Contains(string(contents), "Co-authored-by: Matt Polito <matt.polito@gmail.com>") {
			t.Error("The Co-author line was not included in the commit message")
		}

		if !strings.Contains(string(contents), "abc123") {
			t.Error("The existing contents were replaced")
		}
	})
}
