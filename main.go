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

const version = "1.2.0"

type Config struct {
	Email string
	Devs  map[string]string
}

var opts struct {
	File    string `short:"f" long:"file" description:"Optional alternative git config file"`
	Email   string `short:"e" long:"email" description:"Base author email"`
	Global  bool   `short:"g" long:"global" description:"Modify global git settings"`
	Unset   bool   `short:"u" long:"unset" description:"Unset local pear information"`
	Version bool   `short:"v" long:"version" description:"Print version string"`
}

func pearrcpath() string {
	return path.Join(os.Getenv("HOME"), ".pearrc")
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(user())
		os.Exit(0)
	}

	devs, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		return
	}

	if opts.Version {
		fmt.Printf("Pear version %s\n", version)
		os.Exit(0)
	}

	if opts.Unset {
		git("config", []string{"--remove-section", "user"})
		os.Exit(0)
	}

	sanitizeDevNames(devs)

	var setPairArgs []string
	if opts.File != "" {
		setPairArgs = []string{"--file", opts.File}
	}

	if opts.Global {
		setPairArgs = append(setPairArgs, "--global")
	}

	conf, err := readPearrc(pearrcpath())
	if err != nil {
		log.Fatal(err)
	}

	fullnames := checkPair(devs, conf)
	email := formatEmail(checkEmail(conf), devs)

	setPair(email, fullnames, setPairArgs)

	savePearrc(conf, pearrcpath())
}

func user(args ...string) string {
	options := append(args, []string{"--get", "user.name"}...)

	name := git("config", options)

	return trimNewline(string(name))
}

func email(args ...string) string {
	options := append(args, []string{"--get", "user.email"}...)

	email := git("config", options)

	return trimNewline(string(email))
}

func setPair(email string, pairs []string, args []string) {
	pair := strings.Join(pairs, " and ")

	opts := append(args, "user.name", pair)
	git("config", opts)

	opts = append(args, "user.email", email)
	git("config", opts)
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

func git(subcommand string, opts []string) []byte {
	args := append([]string{subcommand}, opts...)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		log.Print(err)
	}

	return out
}
