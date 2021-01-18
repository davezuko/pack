package api

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	esbuild "github.com/evanw/esbuild/pkg/api"

	"github.com/davezuko/pack/internal/bundler"
	"github.com/davezuko/pack/internal/fs"
	"github.com/davezuko/pack/internal/logger"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func newImpl(opts NewOptions) error {
	if fs.Exists(opts.Path) {
		return fmt.Errorf("destination already exists")
	}

	// Clone template
	err := exec.Command("git", "clone", opts.Template, opts.Path).Run()
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

func startImpl(opts StartOptions) (ServeResult, error) {
	b := bundler.New(bundler.NewOptions{Mode: "development"})
	pkgBundler := bundler.NewPackageBundler()
	sources := http.FileServer(http.Dir(opts.SourceDir))
	statics := http.FileServer(http.Dir(opts.StaticDir))
	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" || !strings.HasPrefix(req.URL.Path, "/") {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte("404 - Not Found"))
			return
		}

		query := path.Clean(req.URL.Path)
		srcPath := path.Join(opts.SourceDir, query)

		// fmt.Printf("Request: %s\n", query)
		if strings.HasPrefix(query, "/web_modules") {
			pkg := query
			pkg = strings.Replace(pkg, "/web_modules/", "", 1)
			pkg = strings.Replace(pkg, ".js", "", 1)
			result := pkgBundler.Bundle(pkg)
			sendBuildResult(res, result)
			return
		}

		// if file does not exist in the source directory, fall back to serving
		// it from the static directory.
		if !fs.Exists(srcPath) {
			statics.ServeHTTP(res, req)
			return
		}

		switch path.Ext(query) {
		case ".ts", ".tsx":
			if opts.Bundle {
				result := b.Bundle([]string{srcPath})
				sendBuildResult(res, result)
			} else {
				result := b.Transform(srcPath)
				sendBuildResult(res, result)
			}
		default:
			sources.ServeHTTP(res, req)
		}
	})
	return newServer(newServerOpts{
		Host:    opts.Host,
		Port:    opts.Port,
		Open:    opts.Open,
		Handler: handler,
	})
}

func sendBuildResult(res http.ResponseWriter, result esbuild.BuildResult) {
	if len(result.OutputFiles) != 1 {
		res.WriteHeader(http.StatusServiceUnavailable)
	} else {
		res.Header().Add("Content-Type", "text/javascript")
		res.Write(result.OutputFiles[0].Contents)
	}
}

func serveImpl(opts ServeOptions) (ServeResult, error) {
	statics := http.Dir(opts.Path)
	handler := http.HandlerFunc(http.FileServer(statics).ServeHTTP)
	return newServer(newServerOpts{
		Host:    opts.Host,
		Port:    opts.Port,
		Open:    opts.Open,
		Handler: handler,
	})
}

func buildImpl(opts BuildOptions) BuildResult {
	log := logger.New()
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	bdl := bundler.New(bundler.NewOptions{
		Mode:   "development",
		Minify: opts.Minify,
	})

	if err := fs.Clean(opts.OutputDir); err != nil {
		log.AddError(fmt.Sprintf("failed to clean output directory: %s", err))
		return loggerToBuildResult(log)
	}
	if fs.Exists(opts.StaticDir) {
		if err := fs.CopyDir(opts.StaticDir, opts.OutputDir); err != nil {
			log.AddError(fmt.Sprintf("failed to copy static directory: %s", err))
		}
	}

	var wg sync.WaitGroup
	filepath.Walk(opts.SourceDir, func(sourceFile string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		switch path.Ext(sourceFile) {
		case ".js", ".ts", ".tsx":
			// noop, these should get bundled
			// TODO: warn on unreferenced scripts?
		case ".html":
			wg.Add(1)
			go func() {
				defer wg.Done()
				entry, err := bundler.NewHTMLEntry(sourceFile)
				if err != nil {
					log.AddError(fmt.Sprintf("failed to load %s: %s", sourceFile, err))
					return
				}
				result, err := entry.Bundle(bundler.HTMLBundleOptions{
					SourceDir: opts.SourceDir,
					Bundler:   bdl,
				})
				if err != nil {
					log.AddError(fmt.Sprintf("failed to build %s: %s", sourceFile, err))
				}
				html := result.HTML
				if opts.Minify {
					minified, err := m.String("text/html", html)
					if err != nil {
						log.AddWarning(fmt.Sprintf("failed to minify %s: %s", sourceFile, err))
					} else {
						html = minified
					}
				}
				outFile, _ := filepath.Rel(opts.SourceDir, sourceFile)
				outFile = path.Join(opts.OutputDir, outFile)
				fmt.Printf("emit: %s\n", outFile)
				ioutil.WriteFile(outFile, []byte(html), 0755)
				for _, file := range result.Scripts {
					outFile := path.Join(opts.OutputDir, file.Path)
					fmt.Printf("emit: %s\n", outFile)
					ioutil.WriteFile(outFile, file.Contents, 0755)
				}
			}()
		default:
			// copy as-is
			wg.Add(1)
			go func() {
				defer wg.Done()
				outFile, _ := filepath.Rel(opts.SourceDir, sourceFile)
				outFile = path.Join(opts.OutputDir, outFile)
				err := fs.CopyFile(sourceFile, outFile)
				if err != nil {
					log.AddError(fmt.Sprintf("could not copy %s: %s", outFile, err))
				} else {
					fmt.Printf("emit: %s\n", outFile)
				}
			}()
		}
		return nil
	})
	wg.Wait()
	return loggerToBuildResult(log)
}

func loggerToBuildResult(log logger.Log) BuildResult {
	result := BuildResult{}
	result.Errors = make([]Message, len(log.Errors()))
	result.Warnings = make([]Message, len(log.Warnings()))
	for i, msg := range log.Errors() {
		result.Errors[i] = Message{Text: msg.Data.Text}
	}
	for i, msg := range log.Warnings() {
		result.Warnings[i] = Message{Text: msg.Data.Text}
	}
	return result
}

type newServerOpts struct {
	Host    string
	Port    uint16
	Open    bool
	Handler http.HandlerFunc
}

func newServer(opts newServerOpts) (ServeResult, error) {
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 3000
	}
	url := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	listener, err := net.Listen("tcp", url)
	if err != nil {
		return ServeResult{}, err
	}

	wait := make(chan error, 1)
	result := ServeResult{
		Host: opts.Host,
		Port: opts.Port,
		Wait: func() error { return <-wait },
		Stop: func() { listener.Close() },
	}
	go func() {
		err := http.Serve(listener, opts.Handler)
		if err != http.ErrServerClosed {
			wait <- err
		} else {
			wait <- nil
		}
	}()
	if opts.Open {
		open(url)
	}
	return result, nil
}

func open(url string) error {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start"}
    case "darwin":
        cmd = "open"
    default:
        cmd = "xdg-open"
    }
    args = append(args, url)
    return exec.Command(cmd, args...).Start()
}