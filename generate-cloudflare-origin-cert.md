# Cloudflare Origin Certificate Setup for Wildcard SSL

To enable HTTPS for all subdomains without browser warnings, you need to:

## 1. Generate Cloudflare Origin Certificate

1. Log in to Cloudflare Dashboard
2. Go to SSL/TLS > Origin Server
3. Click "Create Certificate"
4. Choose:
   - Private key type: RSA
   - Certificate Validity: 15 years (recommended)
   - Hostnames:
     - tunels.tech
     - *.tunels.tech
5. Click "Create"
6. Save the Origin Certificate and Private Key

## 2. Install Certificate on Server

Save the certificate as `/opt/ngrok/certs/cloudflare-origin.crt`
Save the private key as `/opt/ngrok/certs/cloudflare-origin.key`

```bash
# Set proper permissions
sudo chown ngrok:ngrok /opt/ngrok/certs/cloudflare-origin.*
sudo chmod 600 /opt/ngrok/certs/cloudflare-origin.key
sudo chmod 644 /opt/ngrok/certs/cloudflare-origin.crt
```

## 3. Configure Cloudflare SSL/TLS Settings

In Cloudflare Dashboard:
1. Go to SSL/TLS > Overview
2. Set encryption mode to "Full (strict)" or "Full"
3. Go to SSL/TLS > Edge Certificates
4. Ensure "Always Use HTTPS" is ON
5. Ensure "Automatic HTTPS Rewrites" is ON

## 4. DNS Configuration

Ensure these DNS records exist in Cloudflare:
- A record: `tunels.tech` -> 74.235.203.177 (Proxied - Orange Cloud ON)
- CNAME record: `*.tunels.tech` -> tunels.tech (Proxied - Orange Cloud ON)

With Cloudflare proxy enabled (orange cloud), all HTTPS traffic will:
1. Use Cloudflare's trusted SSL certificate for browser connections
2. Use the Origin Certificate for Cloudflare-to-server connections
3. Support unlimited subdomains with wildcard