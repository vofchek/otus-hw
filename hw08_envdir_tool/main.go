package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("set env directory")
	}
	dir := os.Args[1]

	if len(os.Args) <= 2 {
		log.Fatal("set command and params")
	}
	commandWithParams := make([]string, len(os.Args)-2)
	copy(commandWithParams, os.Args[2:])

	env, err := ReadDir(dir)
	if err != nil {
		log.Fatalf("env read error: %v", err)
	}

	code := RunCmd(commandWithParams, env)

	if code != 0 {
		os.Exit(code)
	}
}
