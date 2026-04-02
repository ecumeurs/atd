package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticAssets embed.FS

// GetFileSystem returns an http.FileSystem representing the static assets.
// If devMode is true, it serves from the provided staticPath (filesystem).
// Otherwise, it serves from the embedded assets.
func GetFileSystem(devMode bool, staticPath string) http.FileSystem {
	if devMode {
		return http.Dir(staticPath)
	}

	// The web assets are in the 'static' directory within the embed.FS.
	// We use fs.Sub to serve content from that subdirectory.
	subFS, err := fs.Sub(staticAssets, "static")
	if err != nil {
		// This should never happen if the path is correct
		panic(err)
	}

	return http.FS(subFS)
}

// GetFileContent retrieves a file's content from the appropriate source.
func GetFileContent(fs http.FileSystem, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	content := make([]byte, stat.Size())
	_, err = f.Read(content)
	return content, err
}
