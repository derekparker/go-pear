package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
	"gopkg.in/v1/yaml"
)

const version = "2.1.0.alpha"

type Dev struct {
	Name string
	Email string
}

type Config struct {
	Email string
	Devs  map[string]Dev
}

type options struct {
	Email   string `short:"e" long:"email" description:"Base author email"`
	Unset   bool   `short:"u" long:"unset" description:"Unset local pear information"`
	Version bool   `short:"v" long:"version" description:"Print version string"`
	Debug   bool   `short:"d" long:"debug" description:"Put debug information into git hook"`
}

func pearrcpath() string {
	return path.Join(os.Getenv("HOME"), ".pearrc")
}

func parseFlags() ([]string, *options, error) {
	opts := &options{}
	devs, err := flags.ParseArgs(opts, os.Args[1:])
	if err != nil {
		return nil, nil, err
	}

	return devs, opts, nil
}

func printStderrAndDie(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

func main() {
	devs, opts, err := parseFlags()
	if err != nil {
		return
	}

	if opts.Version {
		fmt.Printf("Pear version %s\n", version)
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		fmt.Println(username())
		os.Exit(0)
	}

	conf, err := readPearrc(pearrcpath())
	if err != nil {
		printStderrAndDie(err)
	}

	sanitizeDevNames(devs)

	if opts.Unset {
		removePair()
		os.Exit(0)
	}

	var (
		devValues  = checkPair(devs, conf)
		email = formatEmail(checkEmail(conf), devs)
	)

	setPair(email, devValues)
	writeHook(email, devValues, opts)
	savePearrc(conf, pearrcpath())
}


func username() string {
	output, err := gitConfig("user.name")
	if err != nil {
		var exitCode int

		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
			if exitCode == 1 {
				log.Fatal("No git user is currently set, try `git config user.name` to confirm")
			} else {
				log.Fatal(output, err)
			}
		}
	}

	return strings.Trim(string(output), "\n ")
}

func email() string {
	output, err := gitConfig("user.email")
	if err != nil {
		log.Fatal(err)
	}

	return strings.Trim(string(output), "\n ")
}

func setPair(email string, pairs []Dev) {

	var fullnames []string

	for _, pair := range pairs {
		fullnames = append(fullnames, pair.Name)
	}
	pair := strings.Join(fullnames, " and ")

	_, err := gitConfig("user.name", pair)
	if err != nil {
		log.Fatal(err)
	}

	_, err = gitConfig("user.email", email)
	if err != nil {
		log.Fatal(err)
	}
}

func writeHook(email string, pairs []Dev, opts *options) {
	var hookBuffer bytes.Buffer
	var debugStatements string

	if opts.Debug {
		debugStatements = `
echo "First Arg: $1"
echo "Second Arg: $2"
echo "Third Arg: $3"
`
	}

	hookBuffer.Write([]byte("#!/bin/sh\n\n"))
	hookBuffer.Write([]byte(debugStatements))
	hookBuffer.Write([]byte("\n"))

	hookBuffer.Write([]byte("addAuthors() {\n"))
	hookBuffer.Write([]byte("  cp $1 /tmp/COMMIT_MSG\n"))
	hookBuffer.Write([]byte("  echo \"\\n\\n\" > $1\n"))

	for _, dev := range pairs {
		hookBuffer.Write([]byte("  echo \"Co-authored-by: "))
		hookBuffer.Write([]byte(dev.Name))
		hookBuffer.Write([]byte(" <"))
		hookBuffer.Write([]byte(dev.Email))
		hookBuffer.Write([]byte(">"))
		hookBuffer.Write([]byte("\""))
		hookBuffer.Write([]byte(" >> $1\n"))
	}

	hookBuffer.Write([]byte("  cat /tmp/COMMIT_MSG >> $1\n"))
	hookBuffer.Write([]byte("}\n"))

	caseStatement := `
case "$2,$3" in
  ,)
    addAuthors $1 ;;
  commit,)
    addAuthors $1 ;;
  *) ;;
esac
`

	hookBuffer.Write([]byte(caseStatement))

        hookPath := prepareCommitHookPath()

	err := ioutil.WriteFile(hookPath, hookBuffer.Bytes(), os.ModeExclusive)
	if err != nil {
		log.Fatal(err)
	}
}

func removePair() {
	_, err := gitConfig("--unset", "user.name")
	if err != nil {
		log.Fatal(err)
	}

	_, err = gitConfig("--unset", "user.email")
	if err != nil {
		log.Fatal(err)
	}
}

func gitConfig(args ...string) (string, error) {
	output, err := exec.Command("git", append([]string{"config"}, args...)...).CombinedOutput()
	return string(output), err
}

func checkEmail(conf *Config) string {
	if conf.Email == "" {
		conf.Email = getEmail()
	}

	return conf.Email
}

func checkPair(pair []string, conf *Config) []Dev {
	var devValues []Dev

	for _, devkey := range pair {
		dev, ok := conf.Devs[devkey]

		if !ok {
			dev = Dev{Name: "", Email: ""}
		}

		if nameok := dev.Name; nameok == "" {
			dev.Name  = getDevName(devkey)
		}

		if emailok := dev.Email; emailok == "" {
			dev.Email = getDevEmail(devkey)
		}

		devValues = append(devValues, dev)

		conf.Devs[devkey] = dev
	}

	return devValues
}

func getDevEmail(devName string) string {
	prompt := fmt.Sprintf("Please enter an email for %s (for github integration use email associated with github):", devName)
	return promptForInput(prompt)
}

func getDevName(devName string) string {
	prompt := fmt.Sprintf("Please enter a full name for %s:", devName)
	return promptForInput(prompt)
}

func getEmail() string {
	return promptForInput("Please provide base author email:")
}

func promptForInput(prompt string) string {
	_, err := fmt.Println(prompt)
	if err != nil {
		log.Fatal(err)
	}

	return readInput()
}

func readInput() string {
	os.Stdin.Seek(0, 0)
	buf := bufio.NewReader(os.Stdin)
	inputString, err := buf.ReadString('\n')
	if err != nil {
		log.Fatal("Could not read from stdin: ", err)
	}

	return trimNewline(inputString)
}

func savePearrc(conf *Config, path string) error {
	contents, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	
	err = ioutil.WriteFile(path, contents, os.ModeExclusive)
	if err != nil {
		return err
	}

	return nil
}

func readPearrc(path string) (*Config, error) {
	conf := &Config{
		Devs: make(map[string]Dev),
	}

	file, err := os.Open(path)
	if err != nil {
		file, err = os.Create(path)
		if err != nil {
			return nil, err
		}
	}

	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func formatEmail(email string, devs []string) string {
	parts := strings.Split(email, "@")
	devlist := strings.Join(devs, "+")
	return fmt.Sprintf("%s+%s@%s", parts[0], devlist, parts[1])
}

func trimNewline(s string) string {
	return strings.TrimSuffix(s, "\n")
}

func sanitizeDevNames(devs []string) {
	for i, dev := range devs {
		devs[i] = strings.ToLower(dev)
	}

	sort.Strings(devs)
}

func prepareCommitHookPath() string{
	output, err := exec.Command("git", "rev-parse",  "--git-dir").CombinedOutput()
	if err != nil {
		log.Fatal("Could not find the git dir", err)
	}

	return (strings.Trim(string(output), "\n") + "/hooks/prepare-commit-msg")
}
