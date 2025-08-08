# HTTPS Tunnel Setup

## Overview
The ngrok server now supports secure HTTPS tunnels with valid SSL certificates. Users can create HTTPS tunnels that browsers will trust without any warnings.

## SSL Certificate Configuration

### Current Setup
- **Let's Encrypt Certificate**: Valid for `tunels.tech` domain
- **Location**: `/opt/ngrok/certs/server.crt` and `server.key`
- **Auto-renewal**: Configured via Certbot

### For Full Wildcard Support (Recommended)
To support HTTPS for all subdomains without individual certificates, configure Cloudflare:

1. **Enable Cloudflare Proxy** (Orange Cloud) for DNS records:
   - A record: `tunels.tech` → 74.235.203.177
   - CNAME record: `*.tunels.tech` → tunels.tech

2. **Configure SSL/TLS Mode** in Cloudflare:
   - Go to SSL/TLS → Overview
   - Set to "Flexible" or "Full"
   - Enable "Always Use HTTPS"
   - Enable "Automatic HTTPS Rewrites"

## Client Usage

### Start an HTTPS Tunnel

#### Using the provided script:
```bash
./start-tunnel-https.sh myapp
```
This creates: `https://myapp.tunels.tech`

#### Manual command:
```bash
./bin/ngrok \
    -config=ngrok.yml \
    -subdomain=myapp \
    -proto=https \
    localhost:8000
```

### Important Notes

1. **Protocol Selection**:
   - Use `-proto=http` for HTTP tunnels (http://subdomain.tunels.tech)
   - Use `-proto=https` for HTTPS tunnels (https://subdomain.tunels.tech)

2. **Browser Trust**:
   - With Cloudflare proxy enabled, all HTTPS connections use Cloudflare's trusted certificates
   - No browser warnings or certificate errors
   - Automatic SSL for all subdomains

3. **Direct Connection** (without Cloudflare):
   - Only works for the main domain (tunels.tech)
   - Subdomains require individual certificates or wildcard certificate

## Server Configuration

The ngrokd server is configured to handle both HTTP (port 80) and HTTPS (port 443) traffic:

```bash
/opt/ngrok/bin/ngrokd \
    -domain=tunels.tech \
    -httpAddr=:80 \
    -httpsAddr=:443 \
    -tunnelAddr=:4443 \
    -tlsCrt=/opt/ngrok/certs/server.crt \
    -tlsKey=/opt/ngrok/certs/server.key
```

## Testing

Test your HTTPS tunnel:
```bash
# Start a local web server
python3 -m http.server 8000

# In another terminal, start the HTTPS tunnel
./start-tunnel-https.sh test

# Access via browser
# https://test.tunels.tech
```

The connection should show as secure with a valid certificate in the browser.