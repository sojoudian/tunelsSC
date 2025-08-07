# 01_add_go_mod.md

## What Changed
- Created new `go.mod` file in the migrated project root
- Set Go version to 1.24
- Defined module path as `github.com/inconshreveable/ngrok`
- Added all required dependencies with their versions

## Why This Change
- Go modules are the standard dependency management system since Go 1.11
- Go 1.24 requires explicit module declaration for modern development
- Enables reproducible builds and better dependency tracking
- Replaces the old GOPATH-based approach used in the original code
- Provides automatic dependency version management

## Code Removed
The original project used GOPATH-based dependency management in the Makefile:
```makefile
export GOPATH:=$(shell pwd)

deps: assets
	go get -tags '$(BUILDTAGS)' -d -v ngrok/...
```

## Code Added
```go
module github.com/inconshreveable/ngrok

go 1.24

require (
	github.com/alecthomas/log4go v0.0.0-20180109082532-d146e6b86faa
	github.com/gorilla/websocket v1.5.3
	github.com/inconshreveable/go-vhost v1.0.0
	github.com/inconshreveable/mousetrap v1.1.0
	github.com/nsf/termbox-go v1.1.1
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
)

require github.com/mattn/go-runewidth v0.0.16 // indirect
```

## Breaking Changes
- Project now requires Go 1.24 or higher to build
- Must use module-aware commands (`go build`, `go test`, etc.) instead of GOPATH-style commands
- Import paths must now use the module path `github.com/inconshreveable/ngrok` instead of just `ngrok`
- The `go-bindata` tool needs to be installed differently (no longer via GOPATH)

## Additional Notes
- The module path `github.com/inconshreveable/ngrok` matches the original repository location
- All dependencies have been updated to their latest compatible versions
- The `github.com/mattn/go-runewidth` is an indirect dependency required by termbox-go
- After creating this file, run `go mod tidy` to ensure all dependencies are properly resolved
- The next step will be to update all import paths in the source code to use the module path
- Some dependencies like `log4go` are quite old and may need replacement in future steps

### Commands to verify the module:
```bash
cd /Users/maziar/Downloads/ngrok/go-1.24-migration/migrated
go mod download  # Download all dependencies
go mod verify    # Verify dependencies
go mod graph     # Show dependency graph
```

### Relevant Go 1.24 Documentation:
- [Go Modules Reference](https://go.dev/ref/mod)
- [Migrating to Go Modules](https://go.dev/blog/migrating-to-go-modules)
- [Module-aware commands](https://go.dev/blog/using-go-modules)