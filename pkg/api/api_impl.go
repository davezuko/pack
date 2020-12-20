package api

import (
	"fmt"

	"github.com/davezuko/pack/internal/build"
	"github.com/davezuko/pack/internal/project"
	"github.com/davezuko/pack/internal/server"
)

func newImpl(opts NewOptions) error {
	if opts.Path == "" {
		return fmt.Errorf("You must provide the name of the project you want to initialize")
	}
	if opts.Template == "" {
		opts.Template = "https://github.com/davezuko/html-template"
	}
	return project.Generate(project.GenerateOpts{
		Path: opts.Path,
		Template: opts.Template,
	})
}

func startImpl(opts StartOptions) (server.ServeResult, error) {
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 3000
	}
	if opts.SourceDir == "" {
		opts.SourceDir = "src"
	}
	if opts.StaticDir == "" {
		opts.StaticDir = "static"
	}
	opts.Bundle = true
	return server.NewDevelopment(server.DevelopmentOpts{
		Port: opts.Port,
		Host: opts.Host,
		Open: opts.Open,
		Bundle: opts.Bundle,
		StaticDir: opts.StaticDir,
		SourceDir: opts.SourceDir,
	})
}

func serveImpl(opts ServeOptions) (server.ServeResult, error) {
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 3000
	}
	return server.NewStatic(server.StaticOpts{
		Port: opts.Port,
		Host: opts.Host,
		Open: opts.Open,
		Path: opts.Path,
	})
}

func buildImpl(opts BuildOptions) (build.BuildResult, error) {
	return build.BuildProject(build.BuildOptions{
		Minify: opts.Minify,
		Hash  : opts.Hash,
		Bundle: opts.Bundle,
		StaticDir: opts.StaticDir,
		SourceDir: opts.SourceDir,
	})
}