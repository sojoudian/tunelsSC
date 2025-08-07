# 03_migrate_deprecated_ioutil.md

## What Changed
- Replaced all usage of the deprecated `io/ioutil` package with modern Go equivalents
- Updated 5 files across client, protocol, and server packages
- Modified import statements and function calls

### Files Modified:
1. `/src/ngrok/client/model.go`
2. `/src/ngrok/client/config.go`
3. `/src/ngrok/proto/http.go`
4. `/src/ngrok/server/tls.go`
5. `/src/ngrok/server/metrics.go`

## Why This Change
- The `io/ioutil` package was deprecated in Go 1.16 and marked for removal
- Go 1.24 continues to support it for compatibility, but using it is considered bad practice
- The functions have been moved to more appropriate packages (`io` and `os`)
- This improves code maintainability and follows Go best practices
- Prepares the codebase for future Go versions where `ioutil` might be removed

## Code Removed

### Import statements removed:
```go
import (
    "io/ioutil"
    // ... other imports
)
```

### Function calls removed:
```go
// In client/model.go
ioutil.ReadAll(localConn)

// In client/config.go
configBuf, err := ioutil.ReadFile(configPath)
oldConfigBytes, err := ioutil.ReadFile(configPath)
err = ioutil.WriteFile(configPath, newConfigBytes, 0600)

// In proto/http.go
return buf.Bytes(), ioutil.NopCloser(buf), err
ioutil.ReadAll(tee.WriteBuffer())
ioutil.ReadAll(tee.ReadBuffer())
return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
req.Body = ioutil.NopCloser(io.LimitReader(neverEnding('x'), req.ContentLength))
ioutil.ReadAll(req.Body)

// In server/tls.go
loadFn := ioutil.ReadFile

// In server/metrics.go
bytes, _ := ioutil.ReadAll(resp.Body)
```

## Code Added

### Updated import statements:
```go
import (
    "io"     // For ReadAll and NopCloser
    "os"     // For ReadFile and WriteFile
    // ... other imports
)
```

### Updated function calls:
```go
// In client/model.go
io.ReadAll(localConn)

// In client/config.go
configBuf, err := os.ReadFile(configPath)
oldConfigBytes, err := os.ReadFile(configPath)
err = os.WriteFile(configPath, newConfigBytes, 0600)

// In proto/http.go
return buf.Bytes(), io.NopCloser(buf), err
io.ReadAll(tee.WriteBuffer())
io.ReadAll(tee.ReadBuffer())
return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
req.Body = io.NopCloser(io.LimitReader(neverEnding('x'), req.ContentLength))
io.ReadAll(req.Body)

// In server/tls.go
loadFn := os.ReadFile

// In server/metrics.go
bytes, _ := io.ReadAll(resp.Body)
```

## Breaking Changes
- None. This is a drop-in replacement with identical functionality
- The migrated functions have the same signatures and behavior
- All existing code continues to work without modification

## Additional Notes

### Migration mapping:
- `ioutil.ReadFile` → `os.ReadFile`
- `ioutil.WriteFile` → `os.WriteFile`
- `ioutil.ReadAll` → `io.ReadAll`
- `ioutil.NopCloser` → `io.NopCloser`
- `ioutil.TempFile` → `os.CreateTemp` (not used in this codebase)
- `ioutil.TempDir` → `os.MkdirTemp` (not used in this codebase)
- `ioutil.ReadDir` → `os.ReadDir` (not used in this codebase)
- `ioutil.Discard` → `io.Discard` (not used in this codebase)

### Best practices applied:
1. Used the most specific package for each function (e.g., file operations in `os`, I/O operations in `io`)
2. Updated all import statements to maintain clean imports
3. Verified that all replacements maintain the same behavior

### Performance considerations:
- No performance impact - these are simple function relocations
- The underlying implementations remain the same

### Testing notes:
- All file I/O operations should be tested to ensure they work correctly
- Pay special attention to error handling, which remains unchanged

### Relevant Go 1.24 Documentation:
- [Go 1.16 Release Notes - Deprecation of io/ioutil](https://go.dev/doc/go1.16#ioutil)
- [io package documentation](https://pkg.go.dev/io)
- [os package documentation](https://pkg.go.dev/os)