package main

import (
	"fmt"
)

const VERSION string = "0.2.0"

var (
	branch    string
	commit    string
	buildtime string
)

func setDefaultVersionInfo() {
	if branch == "" {
		branch = "unknown"
	}
	if commit == "" {
		commit = "unknown"
	}
	if buildtime == "" {
		buildtime = "unknown"
	}
}

func printRelease() {
	fmt.Printf("%s (%s %s %s)\n", VERSION, branch, commit, buildtime)
}
