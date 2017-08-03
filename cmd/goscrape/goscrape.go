package main

import (
	"flag"
	"os"

	"github.com/frizinak/goscrape/cmd"
)

func main() {
	err := cmd.Cmd(flag.CommandLine, os.Args[1:], os.Stderr)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
