package bundler

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	esbuild "github.com/evanw/esbuild/pkg/api"
)

var rewritePackageImports = esbuild.Plugin{
	Name: "rewrite-imports",
	Setup: func(build esbuild.PluginBuild) {
		build.OnResolve(esbuild.OnResolveOptions{
			Filter:    "^[@A-z]",
			Namespace: "file",
		}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
			return esbuild.OnResolveResult{
				Path:      "/web_modules/" + args.Path + ".js",
				Namespace: "module",
				External:  true,
			}, nil
		})
	},
}

var markAllImportsAsExternal = esbuild.Plugin{
	Name: "mark-all-imports-as-external",
	Setup: func(build esbuild.PluginBuild) {
		build.OnResolve(esbuild.OnResolveOptions{
			Filter:    "^.",
			Namespace: "file",
		}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
			if path.Ext(args.Path) == "" {
				args.Path += ".js"
			}
			fmt.Printf("resolve: %s\n", args.Path)
			return esbuild.OnResolveResult{
				Path:     args.Path,
				External: true,
			}, nil
		})
	},
}

type Bundler struct {
	Bundle     func(files []string) esbuild.BuildResult
	Transform  func(file string) esbuild.BuildResult
	BundleDisk func(files []string, outdir string) esbuild.BuildResult // TODO: rethink
}

type NewOptions struct {
	Mode   string
	Minify bool
}

func New(opts NewOptions) Bundler {
	defines := map[string]string{
		"process.env.NODE_ENV": "\"" + opts.Mode + "\"",
	}
	buildOptions := esbuild.BuildOptions{
		Bundle:   true,
		Format:   esbuild.FormatESModule,
		LogLevel: esbuild.LogLevelSilent,
		Define:   defines,
	}
	if opts.Minify {
		buildOptions.MinifySyntax = true
		buildOptions.MinifyWhitespace = true
		buildOptions.MinifyIdentifiers = true
	}
	return Bundler{
		Bundle: func(files []string) esbuild.BuildResult {
			opts := buildOptions
			opts.EntryPoints = files
			return esbuild.Build(opts)
		},
		BundleDisk: func(files []string, outdir string) esbuild.BuildResult {
			opts := buildOptions
			opts.EntryPoints = files
			opts.Outdir = outdir
			return esbuild.Build(opts)
		},
		Transform: func(file string) esbuild.BuildResult {
			opts := buildOptions
			opts.EntryPoints = []string{file}
			opts.Plugins = []esbuild.Plugin{rewritePackageImports, markAllImportsAsExternal}
			return esbuild.Build(opts)
		},
	}
}

type PackageBundler struct {
	Bundle func(pkg string) esbuild.BuildResult
}

func NewPackageBundler() PackageBundler {
	// cache  := make(map[string]esbuild.BuildResult)
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	return PackageBundler{
		Bundle: func(pkg string) esbuild.BuildResult {
			return esbuild.Build(esbuild.BuildOptions{
				Bundle:      true,
				EntryPoints: []string{pkg},
				Format:      esbuild.FormatESModule,
				LogLevel:    esbuild.LogLevelSilent,
				Define:      defines,
				Plugins:     []esbuild.Plugin{rewritePackageImports},
			})
		},
	}
}

type OutputFile struct {
	Path     string
	Contents []byte
}

type BundleHTMLOptions struct {
	Bundler   Bundler
	Path      string
	SourceDir string
	OutputDir string
}

type BundleHTMLResult struct {
	OutputFiles []OutputFile
	Errors      []string
}

func BundleHTML(opts BundleHTMLOptions) (result BundleHTMLResult) {
	result.Errors = []string{}

	f, err := os.Open(opts.Path)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return
	}

	// src/path/to/index.html -> dist/path/to/index.html
	outfile, _ := filepath.Rel(opts.SourceDir, opts.Path)
	outfile = path.Join(opts.OutputDir, outfile)

	nodes := doc.Find("script")
	scripts := make([]string, 0, nodes.Length())
	nodes.Each(func(i int, s *goquery.Selection) {
		uri, _ := s.Attr("src")
		// TODO: join logic incorrect for relative paths
		if strings.HasPrefix(uri, "./") || (strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//")) {
			scripts = append(scripts, path.Join(opts.SourceDir, uri))
			s.Remove()
		}
	})

	nodes = doc.Find("link")
	styles := make([]string, 0, nodes.Length())
	nodes.Each(func(i int, s *goquery.Selection) {
		uri, _ := s.Attr("href")
		// TODO: join logic incorrect for relative paths
		if strings.HasPrefix(uri, "./") || (strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//")) {
			styles = append(styles, path.Join(opts.SourceDir, uri))
			s.Remove()
		}
	})

	entries := make([]string, 0, len(scripts)+len(styles))
	entries = append(entries, scripts...)
	entries = append(entries, styles...)
	// fmt.Printf("entries = %s\n", entries)

	if len(entries) == 0 {
		html, _ := doc.Html()
		result.OutputFiles = []OutputFile{{Path: outfile, Contents: []byte(html)}}
		return
	}

	bundleResult := opts.Bundler.BundleDisk(entries, opts.OutputDir)
	if len(bundleResult.Errors) > 0 {
		for i := range bundleResult.Errors {
			result.Errors = append(result.Errors, bundleResult.Errors[i].Text)
		}
		return
	}

	result.OutputFiles = make([]OutputFile, 0, len(bundleResult.OutputFiles)+1)
	head := doc.Find("head")
	body := doc.Find("body")
	cwd, _ := os.Getwd()
	outdir := path.Join(cwd, opts.OutputDir)
	for _, f := range bundleResult.OutputFiles {
		p, _ := filepath.Rel(outdir, f.Path)
		switch path.Ext(p) {
		case ".css":
			head.AppendHtml(fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\" />", p))
		case ".js":
			// TODO: which scripts should be appended?
			// TODO: script type?
			body.AppendHtml(fmt.Sprintf("<script src=\"%s\"></script>", p))
		}
		result.OutputFiles = append(result.OutputFiles, OutputFile{Path: path.Join(opts.OutputDir, p), Contents: f.Contents})
	}
	html, _ := doc.Html()
	result.OutputFiles = append(result.OutputFiles, OutputFile{Path: outfile, Contents: []byte(html)})
	return
}
