package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/davezuko/pack/pkg/api"
)

type command struct {
	fs   *flag.FlagSet
	Name string
	Run  func(args []string) error
}

func runImpl(args []string) {
	commands := []command{
		buildCommand(),
		newCommand(),
		serveCommand(),
		startCommand(),
	}

	if args[0] == "help" {
		if len(args) == 1 {
			fmt.Printf("\nMissing command for 'pack help <command>'.\n\n")
			fmt.Printf("Available Commands:\n")
			for _, cmd := range commands {
				fmt.Printf("  - %s\n", cmd.Name)
			}
			fmt.Println()
			os.Exit(0)
		}
		for _, cmd := range commands {
			if cmd.Name == args[1] {
				cmd.fs.PrintDefaults()
				os.Exit(0)
			}
		}
	}

	for _, cmd := range commands {
		if cmd.Name == args[0] {
			flgs := []string{}
			argz := []string{}
			for _, arg := range args[1:] {
				if strings.HasPrefix("-", arg) {
					flgs = append(flgs, arg)
				} else {
					argz = append(argz, arg)
				}
			}
			cmd.fs.Parse(flgs)
			err := cmd.Run(argz)
			if err != nil {
				fmt.Printf("\n%s\n\n", err)
				os.Exit(1)
			} else {
				os.Exit(0)
			}
		}
	}

	fmt.Printf(`
Unknown command: "%s".

Tip: run 'pack --help' to see available commands and example usage`, args[0])
	os.Exit(1)
}

func _newCommand(name string) command {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	return command{fs: fs, Name: name}
}

func buildCommand() command {
	cmd := _newCommand("build")

	var bundle bool
	var minify bool
	cmd.fs.BoolVar(&bundle, "bundle", true, "")
	cmd.fs.BoolVar(&minify, "minify", true, "")

	cmd.Run = func(args []string) error {
		result := api.Build(api.BuildOptions{
			SourceDir: "src",
			StaticDir: "static",
			OutputDir: "dist",
			Bundle:    bundle,
			Minify:    minify,
			Hash:      false,
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
	return cmd
}

func newCommand() command {
	cmd := _newCommand("new")

	var template string
	cmd.fs.StringVar(&template, "template", "https://github.com/davezuko/html-template", "")

	cmd.Run = func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("Missing directory name. Try `pack new <directory>`.")
		}
		err := api.New(api.NewOptions{
			Path:     args[0],
			Template: template,
		})
		if err != nil {
			return fmt.Errorf("Something went wrong while creating your project. Sorry about that.\n\n  > %w", err)
		}
		fmt.Printf(`
Success! Your new project is ready to go.

  cd %s && pack start
`, args[0])
		return nil
	}
	return cmd
}

func serveCommand() command {
	cmd := _newCommand("serve")

	var host string
	var port uint
	var open bool
	cmd.fs.StringVar(&host, "host", "localhost", "server host")
	cmd.fs.UintVar(&port, "port", 3000, "server port")
	cmd.fs.BoolVar(&open, "open", false, "automatically open the server")

	cmd.Run = func(args []string) error {
		result, err := api.Serve(api.ServeOptions{
			Path: "dist",
			Host: host,
			Port: uint16(port),
			Open: open,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Server running at %s://%s:%d\n", "http", result.Host, result.Port)
		result.Wait()
		return nil
	}
	return cmd
}

func startCommand() command {
	cmd := _newCommand("start")

	var host string
	var port uint
	var open bool
	cmd.fs.StringVar(&host, "host", "localhost", "server host")
	cmd.fs.UintVar(&port, "port", 3000, "server port")
	cmd.fs.BoolVar(&open, "open", false, "automatically open the server")

	cmd.Run = func(args []string) error {
		result, err := api.Start(api.StartOptions{
			Bundle:    true,
			SourceDir: "src",
			StaticDir: "static",
			Host:      host,
			Port:      uint16(port),
			Open:      open,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Server running at %s://%s:%d\n", "http", result.Host, result.Port)
		result.Wait()
		return nil
	}
	return cmd
}
