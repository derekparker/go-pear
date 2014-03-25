package main

import (
	"log"
	"os/exec"
	"strings"
)

func main() {
	println(globalUser())
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
