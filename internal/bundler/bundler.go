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
	Bundle    func(files []string) esbuild.BuildResult
	Transform func(file string) esbuild.BuildResult
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
		// OutputFiles aren't actually written, so this dir doesn't
		// need to be configurable. We just need a value so that:
		// 1. esbuild can emit > 1 file
		// 2. we can strip the output directory from all OutputFiles
		// before returning to the caller. They never see "dist".
		Outdir: "/dist",
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
			result := esbuild.Build(opts)
			for i := range result.OutputFiles {
				result.OutputFiles[i].Path = strings.TrimPrefix(result.OutputFiles[i].Path, "/dist/")
			}
			return result
		},
		Transform: func(file string) esbuild.BuildResult {
			opts := buildOptions
			opts.EntryPoints = []string{file}
			opts.Plugins = []esbuild.Plugin{rewritePackageImports, markAllImportsAsExternal}
			return esbuild.Build(opts)
		},
	}
}

type OutputFile struct {
	Path     string
	Contents []byte
}

type BundleHTMLOptions struct {
	Bundler Bundler
	Path    string
	Root    string
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

	// <root>/path/to/index.html -> path/to/index.html
	outfile, _ := filepath.Rel(opts.Root, opts.Path)

	// Consume all local scripts and stylesheets from the html document
	entries := []string{}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if uri, ok := s.Attr("src"); ok {
			if uri == "" {
				return // warn?
			}
			switch {
			case strings.HasPrefix(uri, "./"):
				entries = append(entries, path.Join(path.Dir(opts.Path), uri))
				s.Remove()
			case strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//"):
				entries = append(entries, path.Join(opts.Root, uri))
				s.Remove()
			}
		}

	})
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		if uri, ok := s.Attr("href"); ok {
			if uri == "" {
				return // warn?
			}
			switch {
			case strings.HasPrefix(uri, "./"):
				entries = append(entries, path.Join(path.Dir(opts.Path), uri))
				s.Remove()
			case strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//"):
				entries = append(entries, path.Join(opts.Root, uri))
				s.Remove()
			}
		}
	})

	if len(entries) == 0 {
		html, _ := doc.Html()
		result.OutputFiles = []OutputFile{{Path: outfile, Contents: []byte(html)}}
		return
	}

	bundleResult := opts.Bundler.Bundle(entries)
	if len(bundleResult.Errors) > 0 {
		for i := range bundleResult.Errors {
			result.Errors = append(result.Errors, bundleResult.Errors[i].Text)
		}
		return
	}

	result.OutputFiles = make([]OutputFile, 0, len(bundleResult.OutputFiles)+1)
	head := doc.Find("head")
	body := doc.Find("body")
	for _, f := range bundleResult.OutputFiles {
		switch path.Ext(f.Path) {
		case ".css":
			head.AppendHtml(fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\" />", f.Path))
		case ".js":
			// TODO: which scripts should be appended?
			// TODO: consider script type?
			body.AppendHtml(fmt.Sprintf("<script src=\"%s\"></script>", f.Path))
		}
		result.OutputFiles = append(result.OutputFiles, OutputFile{Path: f.Path, Contents: f.Contents})
	}

	html, _ := doc.Html()
	result.OutputFiles = append(result.OutputFiles, OutputFile{Path: outfile, Contents: []byte(html)})
	return
}
