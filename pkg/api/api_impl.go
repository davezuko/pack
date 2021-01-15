package api

import (
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/davezuko/pack/internal/bundler"
	"github.com/davezuko/pack/internal/fs"
)

func newImpl(opts NewOptions) error {
	if opts.Path == "" {
		return fmt.Errorf("Missing project name")
	}
	if opts.Template == "" {
		opts.Template = "https://github.com/davezuko/html-template"
	}
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

	statics := http.FileServer(http.Dir(opts.StaticDir))
	handler := http.HandlerFunc(func (res http.ResponseWriter, req *http.Request) {
		if !(req.Method == "GET" && strings.HasPrefix(req.URL.Path, "/")) {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte("404 - Not Found"))
			return
		}

		query := path.Clean(req.URL.Path)
		if strings.HasSuffix(query, "/") {
			query += "index.html"
		}
		srcPath := path.Join(opts.SourceDir, query)
		// fmt.Printf("Request: %s -> %s\n", query, srcPath)

		// if file does not exist in the source directory, fall back to serving
		// it from the static directory.
		if !fs.Exists(srcPath) {
			statics.ServeHTTP(res, req)
			return
		}

		f, err := os.Open(srcPath)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		switch path.Ext(query) {
		case ".ts", ".tsx", ".js":
			// t1 := time.Now().UnixNano() / int64(time.Millisecond)
			result := bundler.Bundle(srcPath)
			// t2 := time.Now().UnixNano() / int64(time.Millisecond)
			// fmt.Printf("Built %s (%dms)\n", srcPath, t2 - t1)
			if len(result.OutputFiles) != 1 {
				res.WriteHeader(http.StatusInternalServerError)
			} else {
				res.Header().Add("Content-Type", "text/javascript")
				res.Write(result.OutputFiles[0].Contents)
			}
			if len(result.Warnings) > 0 {

			}
			if len(result.Errors) > 0 {

			}
		default:
			f, err := os.Open(srcPath)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer f.Close()
			mt := mime.TypeByExtension(path.Ext(query))
			res.Header().Set("Content-Type", mt)
			res.WriteHeader(http.StatusFound)
			io.Copy(res, f)
		}
	})
	return newServer(newServerOpts{
		Host:    opts.Host,
		Port:    opts.Port,
		Open:    opts.Open,
		Handler: handler,
	})
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
