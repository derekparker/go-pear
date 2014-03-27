package main

import (
	"bufio"
	"fmt"
	"github.com/jessevdk/go-flags"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Config struct {
	Devs map[string]string
}

var opts struct {
	File string `short:"f" long:"file" description:"Optional alternative git config file"`
}

func pearrcpath() string {
	return path.Join(os.Getenv("HOME"), ".pearrc")
}

func main() {
	var setPairArgs []string

	pair, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		log.Fatal("Parse failed: ", err)
	}

	if len(os.Args) == 1 {
		fmt.Println(user())
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		log.Fatal("Must supply 2 arguments")
	}

	if opts.File != "" {
		setPairArgs = []string{"--file", opts.File}
	}

	conf, err := readPearrc(pearrcpath())
	if err != nil {
		log.Fatal(err)
	}

	checkPair(pair, conf)
	fullNamePair := []string{conf.Devs[pair[0]], conf.Devs[pair[1]]}
	setPair(fullNamePair, setPairArgs...)
	savePearrc(conf, pearrcpath())
}

func globalUser() string {
	return user()
}

func user(args ...string) string {
	options := append([]string{"config"}, args...)
	options = append(options, []string{"--get", "user.name"}...)

	cmd := exec.Command("git", options...)
	name, err := cmd.Output()
	if err != nil {
		log.Printf("user lookup failed with: %s", err)
	}

	return trimNewline(string(name))
}

func setPair(pairs []string, args ...string) {
	pair := strings.Join(pairs, " and ")

	options := append([]string{"config"}, args...)
	options = append(options, []string{"user.name", pair}...)

	cmd := exec.Command("git", options...)
	err := cmd.Run()
	if err != nil {
		log.Print(err)
	}
}

func checkPair(pair []string, conf *Config) {
	for _, dev := range pair {
		if _, ok := conf.Devs[dev]; !ok {
			conf.Devs[dev] = getName(dev)
		}
	}
}

func getName(devName string) string {
	_, err := fmt.Printf("Please enter your full name for %s:\n", devName)
	if err != nil {
		log.Fatal(err)
	}

	buf := bufio.NewReader(os.Stdin)
	fullName, err := buf.ReadString('\n')
	if err != nil {
		log.Fatal("Could not read from stdin: ", err)
	}

	return trimNewline(fullName)
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

func trimNewline(s string) string {
	return strings.TrimSuffix(s, "\n")
}
