package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/davezuko/pack/internal/build"
)

type ServeResult struct {
	Addr  func() net.Addr
	Wait  func() error
	Close func()
}

type newOpts struct {
	Host string
	Port uint16
	Open bool
}
func new(handler http.HandlerFunc, opts newOpts) (ServeResult, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port))
	if err != nil {
		// TODO: try to bind to a different port?
		return ServeResult{}, err
	}

	wait := make(chan error, 1)
	result := ServeResult{}
	result.Addr = func() net.Addr { return listener.Addr() }
	result.Wait = func() error { return <-wait }
	result.Close = func() { listener.Close() }
	go func() {
		err := http.Serve(listener, handler)
		if err != http.ErrServerClosed {
			wait <- err
		} else {
			wait <- nil
		}
	}()
	return result, nil
}

type StaticOpts struct {
	Host string
	Port uint16
	Open bool
	Path string
}
func NewStatic(opts StaticOpts) (ServeResult, error) {
	statics := http.Dir(opts.Path)
	handler := http.HandlerFunc(http.FileServer(statics).ServeHTTP)
	return new(handler, newOpts{
		Host: opts.Host,
		Port: opts.Port,
		Open: opts.Open,
	})
}

type DevelopmentOpts struct {
	Host string
	Port uint16
	Open bool
	Bundle bool
	StaticDir string
	SourceDir string
}
func NewDevelopment(opts DevelopmentOpts) (ServeResult, error) {
	staticFS := http.FileServer(http.Dir(opts.StaticDir))
	handler := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/") {
				r.URL.Path += "index.html"
			}
			p := path.Join(opts.SourceDir, path.Clean(r.URL.Path))
			asset, err := build.LoadAsset(p)
			fmt.Printf("asset: %#v\nerror: %s\n", asset, err)
			if err != nil {
				if os.IsNotExist(err) {
					staticFS.ServeHTTP(w, r)
				} else {
					http.Error(w, err.Error(), 500)
				}
				return
			}
			asset.Serve(w, build.AssetServeOpts{Bundle: opts.Bundle})
	})
	return new(handler, newOpts{
		Host: opts.Host,
		Port: opts.Port,
		Open: opts.Open,
	})
}