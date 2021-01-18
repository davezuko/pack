package fs

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Clean(path string) error {
	if Exists(path) {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return os.Mkdir(path, 0755)
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
				return err
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
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

func WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	// TODO: race condition with concurrent writes?
	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, perm)
}