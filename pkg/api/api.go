package api

import (
	"github.com/davezuko/pack/internal/build"
	"github.com/davezuko/pack/internal/server"
)

// NewOptions configures a new project.
type NewOptions struct {
	Path     string
	Template string
}

// ServeOptions configures the static file server.
type ServeOptions struct {
	Host string
	Port uint16
	Open bool
	Path string
}

// StartOptions configures the development server.
type StartOptions struct {
	Host      string
	Port      uint16
	Open      bool
	Bundle    bool
	StaticDir string
	SourceDir string
}

// BuildOptions configures how the project should be built.
type BuildOptions struct {
	Minify    bool
	Hash      bool
	Bundle    bool
	StaticDir string
	SourceDir string
	OutputDir string
}

// Build builds the project to disk and optimizes assets for production.
// The destination folder, specified by options.OutputDir, is self-contained
// and suitable for deployment. For local testing, use the "serve" command
// to serve the output directory.
func Build(opts BuildOptions) (build.BuildResult, error) {
	return buildImpl(opts)
}

// Serve serves the assets that were generated from "build".
func Serve(opts ServeOptions) (server.ServeResult, error) {
	return serveImpl(opts)
}

// Start starts the development server. Assets in opts.SourceDir are built
// on demand. Assets in opts.StaticDir are served without modification.
func Start(opts StartOptions) (server.ServeResult, error) {
	return startImpl(opts)
}

// New creates a new project at the specified path.
func New(opts NewOptions) error {
	return newImpl(opts)
}
