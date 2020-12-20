package cli

import (
	"fmt"
	"os"

	"github.com/davezuko/pack/pkg/api"
)

func runImpl(args []string) {
	var err error
	switch args[0] {
	case "build":
		err = build(args[1:])
	case "new":
		err = new(args[1:])
	case "serve":
		err = serve(args[1:])
	case "start":
		err = start(args[1:])
	default:
		err = fmt.Errorf(`
Unknown command: "%s".

Tip: run 'pack --help' to see available commands and example usage`, args[0])
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func build(args []string) error {
	return nil
}

func new(args []string) error {
	if len(args) == 0 {
		fmt.Printf(`
You must provide the name of the project you want to initialize.

# Example:
pack new <my-project>
`)
		os.Exit(1)
	}
	opts := api.NewOptions{}
	opts.Path = args[0]
	err := api.New(opts)
	if err != nil {
		return err
	}
	return nil
}

func serve(args []string) error {
	opts := api.ServeOptions{}
	api.Serve(opts)
	return nil
}

func start(args []string) error {
	opts := api.StartOptions{}
	srv, err := api.Start(opts)
	if err != nil {
		return err
	}
	fmt.Printf("Server running at %s\n", srv.Addr())
	srv.Wait()
	return nil
}
