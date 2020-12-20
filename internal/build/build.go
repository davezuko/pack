package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

type BuildResult struct {
	Assets []Asset
}

type BuildOptions struct {
	Minify    bool
	Hash      bool
	Bundle    bool
	StaticDir string
	SourceDir string
	OutputDir string
}

func BuildProject(opts BuildOptions) (BuildResult, error) {
	result := BuildResult{Assets: make([]Asset, 0, 10)}
	CopyDir(opts.StaticDir, opts.OutputDir)

	var wg sync.WaitGroup
	filepath.Walk(opts.SourceDir, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			rel, _ := filepath.Rel(opts.SourceDir, path)
			dest := filepath.Join(opts.OutputDir, rel)
			// TODO: notify if dest already exists
			asset, _ := LoadAsset(path)
			err := asset.Build(AssetBuildOpts{
				Minify:  opts.Minify,
				Bundle:  opts.Bundle,
				OutPath: dest,
			})
			if err != nil {
				fmt.Printf("failed to build: %s: %s\n", asset.Path, err)
				return
			}
			asset.Contents = []byte{}
			result.Assets = append(result.Assets, asset)
		}()
		return nil
	})
	wg.Wait()
	return result, nil
}

func Clean(dir string) {
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		os.RemoveAll(dir)
		os.Mkdir(dir, 0777)
	}
}

func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}
	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()
	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func transformTypeScript(source []byte) ([]byte, error) {
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	result := esbuild.Transform(string(source), esbuild.TransformOptions{
		Format:      esbuild.FormatESModule,
		LogLevel:    esbuild.LogLevelSilent,
		Define:      defines,
	})

	// rewrite imports

	return result.Code, nil
}