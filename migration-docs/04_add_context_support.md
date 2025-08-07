# 04_add_context_support.md

## What Changed
- Added context.Context support to major functions for proper cancellation and timeout handling
- Updated connection handling, HTTP operations, and long-running processes
- Modified function signatures to accept context as the first parameter (Go convention)
- Added context propagation through the call stack

### Files Modified:
1. `/src/ngrok/conn/conn.go` - Added context to Dial functions
2. `/src/ngrok/client/model.go` - Added context to control loop and proxy handling
3. `/src/ngrok/server/control.go` - Added context to server control handling
4. `/src/ngrok/proto/interface.go` - Updated Protocol interface
5. `/src/ngrok/proto/http.go` - Added context to HTTP protocol handling

## Why This Change
- Context is the standard Go mechanism for cancellation, deadlines, and request-scoped values
- Enables graceful shutdown of goroutines and network operations
- Prevents resource leaks by ensuring operations can be cancelled
- Improves debugging with context-aware logging
- Required for modern Go applications and many libraries
- Enables timeout control at any level of the call stack

## Code Removed

### Old connection functions without context:
```go
// conn/conn.go
func Dial(addr, typ string, tlsConfig *tls.Config) (Conn, error) {
    // Direct dial without cancellation support
}

// client/model.go
func (c *ClientModel) control() {
    // No context for cancellation
    ctlConn, err := conn.Dial(c.serverAddr, "ctl", c.tlsConfig)
}

// proto/interface.go
type Protocol interface {
    GetName() string
    WrapConn(Conn, interface{}) Conn
}
```

### Old goroutine patterns:
```go
// Long-running operations without cancellation
go func() {
    for {
        // No way to stop this cleanly
        time.Sleep(interval)
        doWork()
    }
}()
```

## Code Added

### Updated connection functions with context:
```go
// conn/conn.go
import "context"

func DialContext(ctx context.Context, addr, typ string, tlsConfig *tls.Config) (Conn, error) {
    dialer := &net.Dialer{}
    netConn, err := dialer.DialContext(ctx, "tcp", addr)
    if err != nil {
        return nil, err
    }
    // ... rest of implementation
}

// Backward compatibility wrapper
func Dial(addr, typ string, tlsConfig *tls.Config) (Conn, error) {
    return DialContext(context.Background(), addr, typ, tlsConfig)
}
```

### Updated client model with context:
```go
// client/model.go
func (c *ClientModel) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            c.Info("Shutting down client: %v", ctx.Err())
            return
        default:
            c.control(ctx)
            // ... reconnection logic
        }
    }
}

func (c *ClientModel) control(ctx context.Context) {
    ctlConn, err := conn.DialContext(ctx, c.serverAddr, "ctl", c.tlsConfig)
    if err != nil {
        return
    }
    defer ctlConn.Close()
    
    // Create child context for this control session
    controlCtx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    // Start heartbeat with context
    c.ctl.Go(func() { c.heartbeat(controlCtx, &lastPong, ctlConn) })
}
```

### Updated protocol interface:
```go
// proto/interface.go
type Protocol interface {
    GetName() string
    WrapConn(context.Context, Conn, interface{}) Conn
}

// proto/http.go
func (h *Http) WrapConn(ctx context.Context, c conn.Conn, connCtx interface{}) conn.Conn {
    tee := conn.NewTee(c)
    
    // Use context for goroutine lifecycle
    go h.readRequests(ctx, tee, lastTxn, connCtx)
    go h.readResponses(ctx, tee, lastTxn)
    
    return tee
}
```

### Context-aware goroutines:
```go
// Heartbeat with context
func (c *ClientModel) heartbeat(ctx context.Context, lastPongAddr *int64, conn conn.Conn) {
    ticker := time.NewTicker(pingInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            conn.Debug("Heartbeat cancelled: %v", ctx.Err())
            return
        case <-ticker.C:
            if err := msg.WriteMsg(conn, &msg.Ping{}); err != nil {
                return
            }
        }
    }
}
```

### Main entry points updated:
```go
// main/ngrok/ngrok.go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    client.MainWithContext(ctx)
}

// client/main.go
func MainWithContext(ctx context.Context) {
    // ... setup code
    NewController().RunWithContext(ctx, config)
}
```

## Breaking Changes
- Function signatures that now require context parameter
- Protocol interface methods need context parameter
- Any custom Protocol implementations must be updated
- External packages calling these functions need updates

### Migration guide for dependent code:
```go
// Old code:
conn, err := conn.Dial(addr, "tcp", tlsConfig)

// New code:
conn, err := conn.DialContext(ctx, addr, "tcp", tlsConfig)
// Or for backward compatibility:
conn, err := conn.Dial(addr, "tcp", tlsConfig) // Uses context.Background()
```

## Additional Notes

### Best practices applied:
1. Context is always the first parameter (Go convention)
2. Created child contexts for sub-operations
3. Proper context cancellation with defer
4. Check ctx.Done() in loops
5. Pass context through the entire call stack

### Context timeout examples added:
```go
// For operations that should timeout
ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
defer cancel()

conn, err := conn.DialContext(ctx, addr, "tcp", tlsConfig)
```

### Error handling with context:
```go
select {
case <-ctx.Done():
    return ctx.Err() // Returns context.Canceled or context.DeadlineExceeded
default:
    // Continue normal operation
}
```

### Testing considerations:
- Tests should use context.WithTimeout to prevent hanging
- Use context.WithCancel in tests for clean shutdown
- Verify that all goroutines exit when context is cancelled

### Performance impact:
- Minimal overhead from context checks
- Improved resource cleanup may actually improve performance
- Prevents goroutine leaks

### Future improvements:
- Add request ID to context for tracing
- Use context for passing logger instances
- Implement context-aware metrics collection

### Relevant Go 1.24 Documentation:
- [context package](https://pkg.go.dev/context)
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Working with Contexts in Go](https://go.dev/blog/context-and-structs)