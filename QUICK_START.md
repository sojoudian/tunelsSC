# Quick Start Guide

## Server is Running
Your ngrok server is already running at `tunnel.tunels.tech` on ports:
- **4443**: Tunnel connections (ngrok clients connect here)
- **8080**: HTTP traffic
- **8443**: HTTPS traffic

## Using the Client

Due to certificate verification in the original ngrok code, you have two options:

### Option 1: Use Official ngrok Client (Recommended for Testing)
Download the official ngrok client from https://ngrok.com/download and configure it:

```bash
# Create config file
cat > ~/.ngrok2/ngrok.yml <<EOF
server_addr: tunnel.tunels.tech:4443
trust_host_root_certs: false
EOF

# Run tunnel
ngrok http -config=~/.ngrok2/ngrok.yml -subdomain=test 8000
```

### Option 2: Modify the Built Client
The client at `bin/ngrok` expects specific certificates. To use it, you would need to:
1. Modify the client code to skip TLS verification
2. Or generate certificates that match the expected root CA

### Option 3: Use a Reverse SSH Tunnel (Alternative)
As a quick alternative, you can use SSH tunneling:

```bash
# Forward local port 8000 to server port 8080
ssh -R 8080:localhost:8000 maziar@74.235.203.177
```

Then access your service at: http://tunnel.tunels.tech:8080

## Testing Your Setup

1. Start a local web server:
   ```bash
   python3 -m http.server 8000
   ```

2. Create a tunnel using one of the methods above

3. Access your service through the tunnel URL

## Next Steps

For production use, you should:
1. Get a proper SSL certificate (e.g., Let's Encrypt)
2. Configure proper authentication
3. Set up firewall rules
4. Monitor server logs: `ssh maziar@74.235.203.177 "sudo journalctl -u ngrokd -f"`