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

type HTMLEntry struct {
	Path     string
	Document *goquery.Document
	Scripts  []string
}

func NewHTMLEntry(path string) (entry HTMLEntry, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return
	}
	return HTMLEntry{Path: path, Document: doc}, nil
}

type HTMLBundleOptions struct {
	Bundler
	SourceDir string
}
type HTMLBundleResult struct {
	HTML    string
	Scripts []OutputFile
}

type OutputFile struct {
	Path     string
	Contents []byte
}

func (h *HTMLEntry) Bundle(opts HTMLBundleOptions) (HTMLBundleResult, error) {
	result := HTMLBundleResult{}

	nodes := h.Document.Find("script")
	scripts := make([]string, 0, nodes.Length())
	nodes.Each(func(i int, s *goquery.Selection) {
		uri, _ := s.Attr("src")
		// TODO: join logic incorrect for relative paths
		if strings.HasPrefix(uri, "./") || (strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//")) {
			scripts = append(scripts, path.Join(opts.SourceDir, uri))
			s.Remove()
		}
	})

	// nodes = h.Document.Find("link")
	// styles := make([]string, 0, nodes.Length())
	// nodes.Each(func(i int, s *goquery.Selection) {
	// 	uri, _ := s.Attr("href")
	// 	// TODO: join logic incorrect for relative paths
	// 	if strings.HasPrefix(uri, "./") || (strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//")) {
	// 		scripts = append(scripts, path.Join(opts.SourceDir, uri))
	// 		s.Remove()
	// 	}
	// })

	if len(scripts) == 0 {
		html, _ := h.Document.Html()
		result.HTML = html
		result.Scripts = []OutputFile{}
		return result, nil
	}
	if len(scripts) > 1 {
		return result, fmt.Errorf("only a single local script is currently supported")
	}

	entries := make([]string, 0, len(scripts))
	entries = append(entries, scripts...)

	bundleResult := opts.Bundle(scripts)
	// fmt.Printf("%#v\n", bundleResult)

	if len(bundleResult.OutputFiles) == 0 {
		html, _ := h.Document.Html()
		result.HTML = html
		result.Scripts = []OutputFile{}
		return result, fmt.Errorf("internal bundler error (no files emitted)")
	}

	src := scripts[0]
	src = strings.Replace(src, path.Ext(src), ".js", 1)
	src, _ = filepath.Rel(opts.SourceDir, src)
	body := h.Document.Find("body")
	body.AppendHtml(fmt.Sprintf("<script src=\"%s\"></script>", src))
	html, _ := h.Document.Html()
	result.HTML = html
	result.Scripts = []OutputFile{
		{Path: src, Contents: bundleResult.OutputFiles[0].Contents},
	}
	return result, nil
}
