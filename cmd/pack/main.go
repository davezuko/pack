package main

import (
	"fmt"
	"os"

	"github.com/davezuko/pack/pkg/cli"
)

var version = "0.1.0"

var helpText = func() string {
	return `
Usage:
  pack <command> [options]

Repository:
  https://github.com/davezuko/pack

Commands:
  new              Create a new project
  start            Start the development server
  build            Build the application to disk
  serve            Serve the built application

Options:
  --port           Set the server port (default: 3000)
  --host           Set the server host (default: localhost)
  --version        Print the current version and exit

Examples:
  # Initialize a new project
  pack new <my-project>

  # Start the development server
  pack start

  # Build your application to disk
  pack build

  # Serve your production build
  pack serve

Run 'pack help <command>' for more information about a specific command.
`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf(helpText())
		os.Exit(0)
	}
	for _, arg := range args {
		switch {
		case arg == "-h", arg == "-help", arg == "--help", arg == "/?":
			fmt.Printf(helpText())
			os.Exit(0)

		case arg == "-v", arg == "--version":
			fmt.Printf("%s\n", version)
			os.Exit(0)
		}
	}
	cli.Run(args)
}
