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

func currentEmail() string {
	return email("--file", "fixtures/test.config")
}

func mockHomeEnv(dir string) {
	os.Setenv("HOME", dir)
}

func closeFile(f *os.File) {
	name := f.Name()
	f.Close()
	os.Remove(name)
}

func createPearrc(t *testing.T, contents []byte) *os.File {
	p := path.Join(os.Getenv("HOME"), ".pearrc")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write(contents)
	if err != nil {
		t.Fatal(err)
	}

	return f
}

func mockStdin(t *testing.T, contents string) *os.File {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmp.WriteString(contents + "\n")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmp.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdin = tmp

	return tmp
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
	tmpstdin := mockStdin(t, "Person B")
	tmp, oldstdout := mockStdout(t)
	pearrc := createPearrc(t, []byte("email: foo@example.com\ndevs:\n  dev1: Full Name A"))
	defer func() {
		cleanupStdout(t, tmp, oldstdout)
		closeFile(tmpstdin)
		closeFile(pearrc)
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

	expectedUser := "Full Name A and Person B"
	if currentUser() != expectedUser {
		t.Errorf("Expected %s got %s", expectedUser, currentUser())
	}

	expectedEmail := "foo+dev1+dev2@example.com"
	if currentEmail() != expectedEmail {
		t.Errorf("Expected %s got %s", expectedEmail, currentEmail())
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
	setPair("foo@example.com", []string{"user1"}, []string{"--file", "fixtures/test.config"})
	expected := "user1"

	if currentUser() != expected {
		t.Errorf("Expected %s got %s", expected, currentUser())
	}
}

func TestSetPairWithTwoDevs(t *testing.T) {
	pair := []string{"user1", "user2"}
	formattedEmail := formatEmail("dev@example.com", pair)
	setPair(formattedEmail, pair, []string{"--file", "fixtures/test.config"})
	expectedUser := "user1 and user2"
	expectedEmail := "dev+user1+user2@example.com"

	if currentUser() != expectedUser {
		t.Errorf("Expected %s got %s", expectedUser, currentUser())
	}

	if currentEmail() != expectedEmail {
		t.Errorf("Expected %s got %s", expectedEmail, currentEmail())
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
