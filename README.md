# ngrok Go 1.24 - Modern Self-Hosted Tunnel Solution

A fully modernized version of ngrok 1.x, updated for Go 1.24 with enhanced security, performance, and maintainability.

## Features

### Core Functionality (Preserved from Legacy)
- **Multi-Protocol Support**: HTTP/HTTPS and TCP tunneling
- **Web Interface**: Inspect traffic via web UI on port 4040
- **Terminal UI**: Real-time tunnel status monitoring
- **Auto-Reconnection**: Exponential backoff for network resilience
- **Proxy Support**: HTTP proxy CONNECT method support
- **Virtual Hosting**: Multiple tunnels on a single server
- **TLS Encryption**: Secure client-server communication

### Modern Enhancements (Go 1.24)
- **Go Modules**: Modern dependency management
- **Context Support**: Proper cancellation and timeout handling
- **TLS 1.3**: Latest security standards
- **Rate Limiting**: Built-in abuse prevention
- **Structured Logging**: Using Go's slog package
- **Generic Cache**: Type-safe LRU cache implementation
- **Native Asset Embedding**: Using go:embed directive

## Quick Start

### Building from Source

```bash
# Clone the repository
cd ~/ngrok-go124

# Download dependencies
go mod download

# Build both client and server
make

# Or build individually
make client  # Creates bin/ngrok
make server  # Creates bin/ngrokd
```

### Running the Server

```bash
# Generate self-signed certificates (for testing)
./setup-ngrok-server.sh

# Run the server
./bin/ngrokd -domain="yourdomain.com" -httpAddr=":80" -httpsAddr=":443"

# Or use systemd service
sudo cp ngrokd.service /etc/systemd/system/
sudo systemctl enable ngrokd
sudo systemctl start ngrokd
```

### Running the Client

```bash
# Create a configuration file
cp ngrok.yml ~/.ngrok

# Edit the configuration
vim ~/.ngrok
# Set: server_addr: "yourdomain.com:4443"

# Start a tunnel
./bin/ngrok http 8080  # Expose local port 8080
./bin/ngrok tcp 22     # Expose SSH
```

## What can I do with ngrok?
- Expose any http service behind a NAT or firewall to the internet on a subdomain
- Expose any tcp service behind a NAT or firewall to the internet on a random port
- Inspect all http requests/responses that are transmitted over the tunnel
- Replay any request that was transmitted over the tunnel

## What is ngrok useful for?
- Temporarily sharing a website that is only running on your development machine
- Demoing an app at a hackathon without deploying
- Developing any services which consume webhooks (HTTP callbacks) by allowing you to replay those requests
- Debugging and understanding any web service by inspecting the HTTP traffic
- Running networked services on machines that are firewalled off from the internet

## Developing on ngrok
[ngrok developer's guide](docs/DEVELOPMENT.md)
