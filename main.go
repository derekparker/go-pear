package main

import (
	"fmt"
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

func pearrcpath() string {
	return path.Join(os.Getenv("HOME"), ".pearrc")
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(user())
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		log.Fatal("Must supply 2 arguments")
	}

	conf, err := readPearrc(pearrcpath())
	if err != nil {
		log.Fatal(err)
	}
	pair := os.Args[1:2]

	checkPair(pair, conf)
	setPair(pair)
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

	return trimNewline(name)
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
	_, err := fmt.Println("Please enter your full name:")
	if err != nil {
		log.Fatal(err)
	}

	fullname, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	return trimNewline(fullname)
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
	conf := &Config{}

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

func trimNewline(s []byte) string {
	return strings.TrimSuffix(string(s), "\n")
}
