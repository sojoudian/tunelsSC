package errors

import (
	"errors"
	"fmt"
)

// Common error types
var (
	// Connection errors
	ErrConnectionFailed = errors.New("connection failed")
	ErrConnectionClosed = errors.New("connection closed")
	ErrTimeout          = errors.New("operation timed out")

	// Authentication errors
	ErrAuthFailed       = errors.New("authentication failed")
	ErrInvalidAuthToken = errors.New("invalid auth token")

	// Tunnel errors
	ErrTunnelNotFound  = errors.New("tunnel not found")
	ErrTunnelInUse     = errors.New("tunnel already in use")
	ErrInvalidProtocol = errors.New("invalid protocol")

	// Configuration errors
	ErrInvalidConfig  = errors.New("invalid configuration")
	ErrConfigNotFound = errors.New("configuration file not found")
)

// ConnectionError represents a connection-related error
type ConnectionError struct {
	Addr string
	Err  error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection to %s failed: %v", e.Addr, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// TunnelError represents a tunnel-related error
type TunnelError struct {
	Name string
	Err  error
}

func (e *TunnelError) Error() string {
	return fmt.Sprintf("tunnel %s error: %v", e.Name, e.Err)
}

func (e *TunnelError) Unwrap() error {
	return e.Err
}
