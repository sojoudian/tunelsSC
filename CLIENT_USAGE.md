# ngrok Client Usage

Your ngrok server is running at `tunnel.tunels.tech`. Here's how to use the client:

## Quick Start

1. **Start a simple HTTP tunnel** (exposes local port 8000):
   ```bash
   ./bin/ngrok -config=ngrok.yml -insecure -subdomain=myapp -proto=http 8000
   ```
   This will create: `http://myapp.tunnel.tunels.tech`

2. **Start a TCP tunnel**:
   ```bash
   ./bin/ngrok -config=ngrok.yml -insecure -proto=tcp 22
   ```

3. **Start multiple tunnels**:
   ```bash
   ./bin/ngrok -config=ngrok.yml -insecure start web ssh
   ```

## Configuration File (ngrok.yml)

The basic configuration is already set up:
```yaml
server_addr: tunnel.tunels.tech:4443
trust_host_root_certs: false
```

You can add tunnel configurations:
```yaml
server_addr: tunnel.tunels.tech:4443
trust_host_root_certs: false

tunnels:
  web:
    subdomain: myapp
    proto: http
    addr: 8000
  
  ssh:
    proto: tcp
    addr: 22
```

## Command Line Options

- `-config`: Path to configuration file
- `-insecure`: Skip certificate verification (required for self-signed certs)
- `-subdomain`: Request a specific subdomain (HTTP/HTTPS only)
- `-proto`: Protocol (http, https, tcp)
- `-authtoken`: Authentication token (if configured on server)

## Examples

### Expose a local web server
```bash
# If you have a web server running on port 3000
./bin/ngrok -config=ngrok.yml -insecure -subdomain=mysite -proto=http 3000
```

### Expose SSH access
```bash
./bin/ngrok -config=ngrok.yml -insecure -proto=tcp 22
```

### Use the start script
```bash
./start-tunnel.sh
```

## Notes

- The `-insecure` flag is required because the server uses a self-signed certificate
- Subdomains are only available for HTTP/HTTPS tunnels
- TCP tunnels get random ports assigned by the server
- The web interface is available at http://localhost:4040 when the client is running

## Troubleshooting

If you see certificate errors, make sure you're using the `-insecure` flag.

If the connection fails, check:
1. DNS resolution: `dig tunnel.tunels.tech` should return `74.235.203.177`
2. Port connectivity: `telnet tunnel.tunels.tech 4443` should connect
3. Server status: The server should be running on port 4443