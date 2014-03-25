package main

import (
	"bytes"
	"log"
	"os/exec"
	"os"
	"strings"
	"gopkg.in/v1/yaml"
	"io/ioutil"
)

type Config struct {
	Devs map[string]string
}

func main() {
	if len(os.Args) == 1 {
		println(user())
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		log.Fatal("Must supply 2 arguments")
	}

	dev1 := os.Args[1]
	dev2 := os.Args[2]
	setPair(dev1, dev2)
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

	return strings.TrimSuffix(string(name), "\n")
}

func setPair(dev1, dev2 string, args ...string) {
	pair := bytes.NewBufferString(dev1)
	if dev2 != "" {
		pair.WriteString(" and " + dev2)
	}

	options := append([]string{"config"}, args...)
	options = append(options, []string{"user.name", pair.String()}...)

	cmd := exec.Command("git", options...)
	err := cmd.Run()
	if err != nil {
		log.Print(err)
	}
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
		return nil, err
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
