package assets

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

// Embed all client assets
//
//go:embed all:client
var clientFS embed.FS

// Asset returns the content of the embedded asset
func Asset(name string) ([]byte, error) {
	// Handle different path formats
	cleanName := strings.TrimPrefix(name, "assets/client/")
	cleanName = strings.TrimPrefix(cleanName, "/")

	// Construct the path
	fullPath := path.Join("client", cleanName)

	return clientFS.ReadFile(fullPath)
}

// AssetFS returns the embedded filesystem
func AssetFS() fs.FS {
	// Return a sub-filesystem rooted at the assets directory
	sub, _ := fs.Sub(clientFS, "client")
	return sub
}
