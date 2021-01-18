package cli

import (
	"fmt"
	"os"

	"github.com/davezuko/pack/pkg/api"
)

func runImpl(args []string) {
	var err error

	if args[0] == "help" {
		if len(args) == 1 {
			fmt.Printf("Missing command name\n")
			os.Exit(1)
		}
		switch args[1] {
		case "build":
			fmt.Printf("%s\n", buildHelp())
		case "new":
			fmt.Printf("%s\n", newHelp())
		case "serve":
			fmt.Printf("%s\n", serveHelp())
		case "start":
			fmt.Printf("%s\n", startHelp())
		default:
			err = fmt.Errorf(`
Unknown command: "%s".
			
Tip: run 'pack --help' to see available commands and example usage`, args[1])
		}
	} else {
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
		}
	}

	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func build(args []string) error {
	result := api.Build(api.BuildOptions{
		SourceDir: "src",
		StaticDir: "static",
		OutputDir: "dist",
		Bundle:    true,
		Minify:    true,
		Hash:      true,
	})
	for _, msg := range result.Warnings {
		fmt.Printf("Warning: %s\n", msg.Text)
	}
	for _, msg := range result.Errors {
		fmt.Printf("Error: %s\n", msg.Text)
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("Encountered %d build error(s).", len(result.Errors))
	}
	fmt.Printf("Run `pack serve` to host your production build locally.\n")
	return nil
}

func buildHelp() string {
	return ""
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
	if opts.Template == "" {
		opts.Template = "https://github.com/davezuko/html-template"
	}
	err := api.New(opts)
	if err != nil {
		return err
	}
	return nil
}

func newHelp() string {
	return ""
}

func serve(args []string) error {
	opts := api.ServeOptions{
		Path: "dist",
	}
	result, err := api.Serve(opts)
	if err != nil {
		return err
	}
	fmt.Printf("Server running at %s://%s:%d\n", "http", result.Host, result.Port)
	result.Wait()
	return nil
}

func serveHelp() string {
	return ""
}

func start(args []string) error {
	opts := api.StartOptions{
		Bundle:    true,
		SourceDir: "src",
		StaticDir: "static",
	}
	result, err := api.Start(opts)
	if err != nil {
		return err
	}
	fmt.Printf("Server running at %s://%s:%d\n", "http", result.Host, result.Port)
	result.Wait()
	return nil
}

func startHelp() string {
	return ""
}
