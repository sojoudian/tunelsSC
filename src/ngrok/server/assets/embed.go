package assets

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

// Embed all server assets
//
//go:embed all:server
var serverFS embed.FS

// Asset returns the content of the embedded asset
func Asset(name string) ([]byte, error) {
	// Handle different path formats
	cleanName := strings.TrimPrefix(name, "assets/server/")
	cleanName = strings.TrimPrefix(cleanName, "/")

	// Construct the path
	fullPath := path.Join("server", cleanName)

	return serverFS.ReadFile(fullPath)
}

// AssetFS returns the embedded filesystem
func AssetFS() fs.FS {
	// Return a sub-filesystem rooted at the assets directory
	sub, _ := fs.Sub(serverFS, "server")
	return sub
}
