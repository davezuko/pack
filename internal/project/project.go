package project

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

type GenerateOpts struct {
	Path string
	Template string
}
func Generate(opts GenerateOpts) error {
	if opts.Path == "" {
		return fmt.Errorf("destination not specified")
	}

	// Make sure the destination doesn't already exist
	_, err := os.Stat(opts.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// OK
		} else {
			return err
		}
	} else {
		return fmt.Errorf("destination already exists")
	}

	// Clone template
	err = exec.Command("git", "clone", opts.Template, opts.Path).Run()
	if err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}
	err = os.RemoveAll(path.Join(opts.Path, ".git"))
	if err != nil {
		return err
	}

	// Install dependencies
	cmd := exec.Command("npm", "install")
	cmd.Dir = opts.Path
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %s", err)
	}
	return nil
}