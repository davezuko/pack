package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

var m = minify.New()

var minifiableMediaTypes = map[string]string{
	".css":  "text/css",
	".html": "text/html",
	".js":   "text/javascript",
	".json": "application/json",
	".mjs":  "text/javascript",
}

func init() {
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
}

type Asset struct {
	Path     string
	Contents []byte
}

func LoadAsset(file string) (Asset, error) {
	if _, err := os.Stat(file); err != nil {
		return Asset{}, err
	}
	return Asset{Path: file}, nil
}

type AssetBuildOpts struct {
	Minify bool
	Bundle bool
	OutPath string
}
func (a *Asset) Build(opts AssetBuildOpts) error {
	if opts.Minify {
		mediaType := minifiableMediaTypes[path.Ext(a.Path)]
		if mediaType != "" {
			dat, err := m.Bytes(mediaType, a.Contents)
			if err != nil {
				return fmt.Errorf("minification error: %s", err)
			}
			a.Contents = dat
		}
	}
	return nil
}

type AssetServeOpts struct {
	Bundle bool
}

func (a *Asset) Serve(w http.ResponseWriter, opts AssetServeOpts) {
	ct := a.contentType()
	w.Header().Add("Content-Type", ct)
	
	switch path.Ext(a.Path) {
	case ".ts", ".tsx":
		result, _ := transformTypeScript(a.source())
		w.Write(result)
	default:
		a.write(w)
	}
}

func (a *Asset) source() []byte {
	dat, _ := ioutil.ReadFile(a.Path)
	return dat
}

func (a *Asset) write(w io.Writer) {
	if a.Contents == nil {
		f, _ := os.Open(a.Path)
		io.Copy(w, f)
	} else {
		w.Write(a.Contents)
	}
}

func (a *Asset) contentType() string {
	return mime.TypeByExtension(path.Ext(a.Path))
}
