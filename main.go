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
	"regexp"
	"sort"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
	"gopkg.in/v1/yaml"
)

const version = "2.1.3.alpha"

type Dev struct {
	Name  string
	Email string
}

type Config struct {
	Email string
	Devs  map[string]Dev
}

type options struct {
	Unset       bool   `short:"u" long:"unset" description:"Unset local pear information"`
	Version     bool   `short:"v" long:"version" description:"Print version string"`
	Augment     bool   `short:"a" long:"augment-commit-message" description:"Used within the git hook to write Co-authors to commit message"`
	Integration string `short:"i" long:"github-integration" description:"Takes values on or off, to turn on or off github co-author integration"`
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

func inGitRepository() bool {
	_, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").CombinedOutput()
	return err == nil
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

	if !inGitRepository() {
		fmt.Println("Pear only works in a git repository")
		os.Exit(1)
	}

	if opts.Integration != "" {
		switch opts.Integration {
		case "on":
			writeHook()
			gitConfig("pear.githubIntegration", "true")
		case "off":
			removeHook()
			gitConfig("pear.githubIntegration", "false")
		default:
			fmt.Println("Integration options must be either 'on' or 'off'")
		}
		os.Exit(0)
	}

	if opts.Unset {
		removePair()
		removeHook()
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

	if opts.Augment {
		var revision, source string

		if len(devs) == 0 {
			log.Fatal("A file name as the first argument is required for Augment")
		}

		if len(devs) < 3 {
			revision = ""
		} else {
			revision = devs[2]
		}

		if len(devs) < 2 {
			source = ""
		} else {
			source = devs[1]
		}

		augmentCommitMessage(devs[0], source, revision, conf)
		os.Exit(0)
	}

	sanitizeDevNames(devs)

	var (
		devValues = checkPair(devs, conf)
		email     = formatEmail(checkEmail(conf), devs)
	)

	setPair(email, devValues, devs)
	if githubIntegration() {
		writeHook()
	}
	savePearrc(conf, pearrcpath())
}

func githubIntegration() bool {
	value, _ := gitConfig("pear.githubIntegration")
	return strings.Trim(value, "\n ") != "false"
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

func setPair(email string, pairs []Dev, devs []string) {

	var fullnames []string

	for _, pair := range pairs {
		fullnames = append(fullnames, pair.Name)
	}
	pair := strings.Join(fullnames, " and ")

	_, err := gitConfig("pear.devs", strings.Join(devs, ","))
	if err != nil {
		log.Fatal(err)
	}

	_, err = gitConfig("user.name", pair)
	if err != nil {
		log.Fatal(err)
	}

	_, err = gitConfig("user.email", email)
	if err != nil {
		log.Fatal(err)
	}
}

func getCurrentDevValues(conf *Config) []Dev {
	var devValues []Dev
	devs, err := gitConfig("pear.devs")

	if err != nil {
		fmt.Println("Could not read git config for pairs")
		log.Fatal(err)
	}

	for _, devkey := range strings.Split(devs, ",") {
		dev, ok := conf.Devs[strings.Trim(devkey, "\n ")]
		if !ok {
			log.Fatal("No dev found: " + devkey)
		}

		devValues = append(devValues, dev)
	}

	return devValues
}

func removeHook() {
	var hookBuffer bytes.Buffer

	hookPath := prepareCommitHookPath()

	var contents []byte
	var err error

	contents, err = ioutil.ReadFile(hookPath)
	if err != nil {
		contents = []byte("")
	}

	re := regexp.MustCompile("(?m)[\r\n]+^.*pear.*$")
	replacedString := re.ReplaceAllString(string(contents), "")

	hookBuffer.Write([]byte(replacedString))

	err = ioutil.WriteFile(hookPath, hookBuffer.Bytes(), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func writeHook() {
	var hookBuffer bytes.Buffer

	hookPath := prepareCommitHookPath()

	var contents []byte
	var err error

	contents, err = ioutil.ReadFile(hookPath)
	if err != nil {
		contents = []byte("")
	}

	re := regexp.MustCompile("(?m)[\r\n]+^.*pear.*$")
	replacedString := re.ReplaceAllString(string(contents), "")

	hookBuffer.Write([]byte(replacedString))
	hookBuffer.Write([]byte("\npear --augment-commit-message $1 $2 $3\n"))

	err = ioutil.WriteFile(hookPath, hookBuffer.Bytes(), 0755)
	if err != nil {
		log.Fatal(err)
	}

	if err = os.Chmod(hookPath, 0755); err != nil {
		log.Fatal(err)
	}
}

func removePair() {
	_, err := gitConfig("--unset", "user.name")
	if err != nil {
		if err.Error() != "exit status 5" {
			log.Fatal(err)
		}
	}

	_, err = gitConfig("--unset", "user.email")
	if err != nil {
		if err.Error() != "exit status 5" {
			log.Fatal(err)
		}
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
			dev.Name = getDevFullName(devkey)
		}

		if emailok := dev.Email; emailok == "" && githubIntegration() {
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

func getDevFullName(devName string) string {
	var devFullName string

	for devFullName == "" {
		prompt := fmt.Sprintf("Please enter a full name for %s:", devName)
		devFullName = promptForInput(prompt)
	}

	return devFullName
}

func getEmail() string {
	var baseAuthorEmail string

	re := regexp.MustCompile("^[^@]+@[^@]+$")

	for baseAuthorEmail == "" || !re.MatchString(baseAuthorEmail) {

		baseAuthorEmail = promptForInput("Please provide base author email:")

		if !re.MatchString(baseAuthorEmail) {
			fmt.Println("Invalid")
		}
	}

	return baseAuthorEmail
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

	err = ioutil.WriteFile(path, contents, 0644)
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

func prepareCommitHookPath() string {
	output, err := exec.Command("git", "rev-parse", "--git-dir").CombinedOutput()
	if err != nil {
		log.Fatal("Could not find the git dir", err)
	}

	return (strings.Trim(string(output), "\n") + "/hooks/prepare-commit-msg")
}

func augmentCommitMessage(filePath string, source string, revision string, conf *Config) {
	pairs := getCurrentDevValues(conf)

	switch source + "," + revision {
	case ",":
		addCoauthorsToCommitMessage(filePath, pairs)
	case "commit,":
		addCoauthorsToCommitMessage(filePath, pairs)
	}
}

func addCoauthorsToCommitMessage(filePath string, pairs []Dev) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var commitMessageBuffer bytes.Buffer
	commitMessageBuffer.Write([]byte("\n\n"))

	for _, dev := range pairs {
		commitMessageBuffer.Write([]byte("Co-authored-by: "))
		commitMessageBuffer.Write([]byte(dev.Name))
		commitMessageBuffer.Write([]byte(" <"))
		commitMessageBuffer.Write([]byte(dev.Email))
		commitMessageBuffer.Write([]byte(">\n"))
	}

	commitMessageBuffer.Write(contents)

	ioutil.WriteFile(filePath, commitMessageBuffer.Bytes(), 0644)
}
