#!/bin/bash
# Script to complete ngrok server setup

set -e

echo "ngrok Server Setup Script"
echo "========================"

# Check if running on the remote server
if [ ! -d "/opt/ngrok" ]; then
    echo "Error: ngrok not found at /opt/ngrok. Make sure you're running this on the remote server."
    exit 1
fi

# Step 1: Create configuration directory
echo "Creating configuration directory..."
sudo mkdir -p /etc/ngrok

# Step 2: Create certificate directory
echo "Creating certificate directory..."
sudo mkdir -p /opt/ngrok/certs

# Step 3: Generate self-signed certificates (for testing)
echo "Generating self-signed TLS certificates..."
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /opt/ngrok/certs/server.key \
    -out /opt/ngrok/certs/server.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=tunnel.example.com"

# Step 4: Set proper permissions
echo "Setting certificate permissions..."
sudo chown -R ngrok:ngrok /opt/ngrok/certs
sudo chmod 600 /opt/ngrok/certs/server.key
sudo chmod 644 /opt/ngrok/certs/server.crt

# Step 5: Create basic configuration file
echo "Creating configuration file..."
sudo tee /etc/ngrok/ngrokd.yml > /dev/null <<EOF
# ngrok server configuration
server_addr: "0.0.0.0:4443"
http_addr: "0.0.0.0:80"
https_addr: "0.0.0.0:443"
domain: "tunnel.example.com"
tls_cert: "/opt/ngrok/certs/server.crt"
tls_key: "/opt/ngrok/certs/server.key"
log_to: "/var/log/ngrok/ngrokd.log"
log_level: "INFO"
EOF

# Step 6: Create log directory
echo "Creating log directory..."
sudo mkdir -p /var/log/ngrok
sudo chown ngrok:ngrok /var/log/ngrok

# Step 7: Set configuration permissions
echo "Setting configuration permissions..."
sudo chown ngrok:ngrok /etc/ngrok/ngrokd.yml
sudo chmod 644 /etc/ngrok/ngrokd.yml

echo ""
echo "Setup completed! Next steps:"
echo "1. Edit /etc/ngrok/ngrokd.yml and update:"
echo "   - domain: Set to your actual domain name"
echo "   - Replace self-signed certs with real ones if needed"
echo ""
echo "2. Start the service:"
echo "   sudo systemctl start ngrokd"
echo ""
echo "3. Enable auto-start:"
echo "   sudo systemctl enable ngrokd"
echo ""
echo "4. Check service status:"
echo "   sudo systemctl status ngrokd"
echo ""
echo "5. View logs:"
echo "   sudo journalctl -u ngrokd -f"