package api

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	esbuild "github.com/evanw/esbuild/pkg/api"

	"github.com/davezuko/pack/internal/bundler"
	"github.com/davezuko/pack/internal/fs"
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
	if opts.SourceDir == "" {
		opts.SourceDir = "src"
	}
	if opts.StaticDir == "" {
		opts.StaticDir = "static"
	}

	sources := http.FileServer(http.Dir(opts.SourceDir))
	statics := http.FileServer(http.Dir(opts.StaticDir))
	pkgBundler := bundler.NewPackageBundler()
	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if !(req.Method == "GET" && strings.HasPrefix(req.URL.Path, "/")) {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte("404 - Not Found"))
			return
		}

		query := path.Clean(req.URL.Path)
		srcPath := path.Join(opts.SourceDir, query)

		fmt.Printf("Request: %s\n", query)
		if strings.HasPrefix(query, "/web_modules") {
			pkg := query
			pkg = strings.Replace(pkg, "/web_modules/", "", 1)
			pkg = strings.Replace(pkg, ".js", "", 1)
			result := pkgBundler.Build(pkg)
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
				result := bundler.Bundle(srcPath)
				sendBuildResult(res, result)
			} else {
				result := bundler.Transform(srcPath)
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
	if opts.Path == "" {
		opts.Path = "dist"
	}
	statics := http.Dir(opts.Path)
	handler := http.HandlerFunc(http.FileServer(statics).ServeHTTP)
	return newServer(newServerOpts{
		Host:    opts.Host,
		Port:    opts.Port,
		Open:    opts.Open,
		Handler: handler,
	})
}

func buildImpl(opts BuildOptions) (BuildResult, error) {
	fs.Clean(opts.OutputDir)
	fs.CopyDir(opts.StaticDir, opts.OutputDir)

	var wg sync.WaitGroup
	filepath.Walk(opts.SourceDir, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
		return nil
	})
	wg.Wait()
	return BuildResult{}, nil
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
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port))
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
	return result, nil
}
