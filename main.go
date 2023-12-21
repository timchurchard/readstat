package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/timchurchard/readstat/cmd"
)

const cliName = "readstat"

func main() {
	if len(os.Args) < 2 {
		usageRoot()
	}

	// Save the command and reset the flags
	command := os.Args[1]
	flag.CommandLine = flag.NewFlagSet(cliName, flag.ExitOnError)
	os.Args = append([]string{cliName}, os.Args[2:]...)

	switch command {
	case "sync":
		os.Exit(cmd.Sync(os.Stdout))
	}

	usageRoot()
}

func usageRoot() {
	fmt.Printf("usage: %s command(sync) options\n", cliName)
	os.Exit(1)
}
