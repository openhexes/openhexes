package server

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed ui
var uiFS embed.FS

type FallbackFS struct {
	fs.FS
	fallback string
}

func NewFallbackFS(f fs.FS, fallback string) *FallbackFS {
	return &FallbackFS{
		FS:       f,
		fallback: fallback,
	}
}

func (f *FallbackFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		return f.FS.Open(f.fallback)
	}
	return file, err
}

func GetUIHandler() (http.Handler, error) {
	fs, err := fs.Sub(uiFS, "ui")
	if err != nil {
		return nil, fmt.Errorf("building sub-fs: %w", err)
	}
	return http.FileServerFS(NewFallbackFS(fs, "index.html")), nil
}
