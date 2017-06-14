package fileutil

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/oakmoundstudio/oak/dlog"
)

var (
	// BindataFn is a function to access binary data outside of os.Open
	BindataFn  func(string) ([]byte, error)
	BindataDir func(string) ([]string, error)
	wd, _      = os.Getwd()
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

// Open is a wrapper around os.Open that will also check a function to access
// byte data. The intended use is to use the go-bindata library to create an
// Asset function that matches this signature.
func Open(file string) (io.ReadCloser, error) {
	// Check bindata
	if BindataFn != nil {
		// It looks like we need to clean this output sometimes--
		// we get capitalization where we don't want it ocassionally?
		rel, err := filepath.Rel(wd, file)
		if err == nil {
			data, err := BindataFn(rel)
			if err == nil {
				dlog.Verb("Found file in binary,", rel)
				// convert data to io.Reader
				return nopCloser{bytes.NewReader(data)}, err
			}
			dlog.Warn("File not found in binary", rel)
		} else {
			dlog.Warn("Error in rel", err)
		}
	}
	return os.Open(file)
}

func ReadFile(file string) ([]byte, error) {
	if BindataFn != nil {
		rel, err := filepath.Rel(wd, file)
		if err == nil {
			return BindataFn(rel)
		}
		dlog.Warn("Error in rel", err)
	}
	return ioutil.ReadFile(file)
}

func ReadDir(file string) ([]os.FileInfo, error) {
	var fis []os.FileInfo
	if BindataDir != nil {
		rel, err := filepath.Rel(wd, file)
		if err == nil {
			strs, err := BindataDir(rel)
			if err == nil {
				fis := make([]os.FileInfo, len(strs))
				for i, s := range strs {
					// If the data does not contain a period, we consider it
					// a directory
					fis[i] = dummyfileinfo{s, !strings.ContainsRune(s, '.')}
				}
			} else {
				dlog.Warn(err)
			}
		} else {
			dlog.Warn(err)
		}
	}
	fis2, err := ioutil.ReadDir(file)
	if err != nil {
		return fis, err
	}
	return append(fis, fis2...), nil
}
