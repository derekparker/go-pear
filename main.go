package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
	"gopkg.in/v1/yaml"
)

const version = "1.3.2"

type Config struct {
	Email string
	Devs  map[string]string
}

type options struct {
	File    string `short:"f" long:"file" description:"Optional alternative git config file"`
	Email   string `short:"e" long:"email" description:"Base author email"`
	Global  bool   `short:"g" long:"global" description:"Modify global git settings"`
	Unset   bool   `short:"u" long:"unset" description:"Unset local pear information"`
	Version bool   `short:"v" long:"version" description:"Print version string"`
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
		fullnames = checkPair(devs, conf)
		email     = formatEmail(checkEmail(conf), devs)
	)

	setPair(email, fullnames)
	savePearrc(conf, pearrcpath())
}

func username() string {
	name, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		log.Fatal(err)
	}

	return strings.Trim(string(name), "\n ")
}

func email() string {
	email, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		log.Fatal(err)
	}

	return strings.Trim(string(email), "\n ")
}

func setPair(email string, pairs []string) {
	pair := strings.Join(pairs, " and ")

	// git config user.name <value>
	cmd := exec.Command("git", "config", "user.name", pair)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("git", "config", "user.email", email)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func removePair() {
	cmd := exec.Command("git", "config", "--unset", "user.name")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("git", "config", "--unset", "user.email")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func checkEmail(conf *Config) string {
	if conf.Email == "" {
		conf.Email = getEmail()
	}

	return conf.Email
}

func checkPair(pair []string, conf *Config) []string {
	var fullnames []string
	for _, dev := range pair {
		if _, ok := conf.Devs[dev]; !ok {
			conf.Devs[dev] = getName(dev)
		}

		fullnames = append(fullnames, conf.Devs[dev])
	}

	return fullnames
}

func getName(devName string) string {
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
		Devs: make(map[string]string),
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
