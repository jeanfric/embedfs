// Package embedfs provides an http.FileSystem implementation backed
// by a map[string]string.
//
// It is meant to be used with the goembed tool,
// cf. https://github.com/jeanfric/goembed.
package embedfs

import (
	"net/http"
	"os"
	"strings"
	"time"
)

type embedFS struct {
	files map[string]string
	dirs  map[string][]os.FileInfo
}

type embedFile struct {
	name   string
	reader *strings.Reader
	len    int64
	isDir  bool
	fis    []os.FileInfo
}

type embedFileInfo struct {
	len   int64
	name  string
	isDir bool
}

// New returns an http.FileSystem constructed from m, a map of string
// paths to string contents.  The m map must not be modified after
// calling New.  The map keys must use slash ('/') as path separator,
// and all paths must being with '/'.  For example:
//
//	m := make(map[string]string)
//	m["/index.html"] = "<html><head>..."
//	m["/javascript/app.js"] = "..."
//	m["/images/gopher.png"] = "..."
//
// See also: the github.com/jeanfric/goembed/cmd/goembed command can
// be used to embed file assets into a map that can be passed directly
// to New.  Refer to https://github.com/jeanfric/goembed.
func New(m map[string]string) http.FileSystem {
	// Be done with computing the directories right away.
	// We don't expect the underlying map to change.
	dirs := make(map[string][]os.FileInfo)

	// Map keys used as set, to uniquify the directory list.
	for k, _ := range m {
		dirs[dirname(k)] = nil
	}

	// For each known directory, compute its contents.
	for d, _ := range dirs {
		fis := make([]os.FileInfo, 0)

		for k, v := range m {
			if dirname(k) == d {
				fis = append(fis, &embedFileInfo{
					len:   int64(len(v)),
					name:  basename(k),
					isDir: false,
				})
			}
		}

		for k, _ := range dirs {
			// Skip recording '/' as containing '/' with
			// the 'k != d' test.
			if k != d && dirname(k) == d {
				fis = append(fis, &embedFileInfo{
					len:   0,
					name:  basename(k),
					isDir: true,
				})
			}
		}
		dirs[d] = fis
	}

	return &embedFS{
		files: m,
		dirs:  dirs,
	}
}

func (fs *embedFS) Open(name string) (http.File, error) {
	s, ok := fs.files[name]
	if !ok {
		_, ok := fs.dirs[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return &embedFile{
			reader: nil,
			name:   name,
			len:    0,
			isDir:  true,
			fis:    fs.dirs[name],
		}, nil

	}
	return &embedFile{
		reader: strings.NewReader(s),
		name:   name,
		len:    int64(len(s)),
	}, nil
}

func (f *embedFile) Stat() (os.FileInfo, error) {
	return &embedFileInfo{
		len:   f.len,
		name:  f.name,
		isDir: f.isDir,
	}, nil
}

func (f *embedFile) Seek(offset int64, whence int) (ret int64, err error) {
	return f.reader.Seek(offset, whence)
}

func (f *embedFile) Close() error { return nil }

func (f *embedFile) IsDir() bool { return f.isDir }

func (f *embedFile) Read(p []byte) (n int, err error) { return f.reader.Read(p) }

func (f *embedFile) Readdir(count int) ([]os.FileInfo, error) { return f.fis, nil }

func (fi *embedFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0700
	}
	return 0400
}
func (fi *embedFileInfo) Sys() interface{} { return nil }

func (fi *embedFileInfo) IsDir() bool { return fi.isDir }

func (fi *embedFileInfo) ModTime() time.Time { return time.Time{} }

func (fi *embedFileInfo) Size() int64 { return fi.len }

func (fi *embedFileInfo) Name() string { return fi.name }

// Dirname removes the last part (rightmost) of the path, returning
// the containing directory part.  If the path is '/', it returns '/'.
func dirname(f string) string {
	i := len(f) - 1
	for f[i] != '/' && i >= 0 {
		i--
	}
	if i < 1 {
		return "/"
	}
	return f[0:i]
}

// Basename removes the directory part of f, returning just the last
// (rightmost) part of the path.  If f is '/', it returns '/'.
func basename(f string) string {
	if f == "/" {
		return "/"
	}
	i := len(f) - 1
	for f[i] != '/' && i > 0 {
		i--
	}
	return f[i+1:]
}
