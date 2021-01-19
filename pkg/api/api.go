package api

// NewOptions configures a new project.
type NewOptions struct {
	Path     string
	Template string
	Yarn     bool
}

// ServeOptions configures the static file server.
type ServeOptions struct {
	Host string
	Port uint16
	Open bool
	Path string
}

// ServeResult holds an active HTTP server.
type ServeResult struct {
	Host string
	Port uint16
	Wait func() error
	Stop func()
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

// BuildResult provides diagnostic information about a build.
type BuildResult struct {
	Errors      []Message
	Warnings    []Message
	OutputFiles []OutputFile
}

type OutputFile struct {
	Path string
}

type Message struct {
	Text string
}

// Build builds the project to options.OutputDir and optimizes assets for
// production. The output directory will be a self-contained application
// and suitable for deployment to a static CDN.
func Build(opts BuildOptions) BuildResult {
	return buildImpl(opts)
}

// Serve serves the assets that were generated from "build".
func Serve(opts ServeOptions) (ServeResult, error) {
	return serveImpl(opts)
}

// Start starts the development server. Assets in opts.SourceDir are built
// on demand. Assets in opts.StaticDir are served without modification.
func Start(opts StartOptions) (ServeResult, error) {
	return startImpl(opts)
}

// New creates a new project at the specified path.
func New(opts NewOptions) error {
	return newImpl(opts)
}
