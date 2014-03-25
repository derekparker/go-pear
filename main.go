package main

import (
	"bytes"
	"log"
	"os/exec"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
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
