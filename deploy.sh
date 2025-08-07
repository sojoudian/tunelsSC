#!/bin/bash
# Deploy script for ngrok Go 1.24 migrated version

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}ngrok Go 1.24 Deployment Script${NC}"
echo "================================="

# Check if we're in the migrated directory
if [ ! -f "go.mod" ] || [ ! -d "src/ngrok" ]; then
    echo -e "${RED}Error: This script must be run from the go-1.24-migration/migrated directory${NC}"
    exit 1
fi

# Parse command line arguments
REMOTE_HOST=""
REMOTE_USER=""
REMOTE_PATH="/tmp/ngrok-migrated"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            REMOTE_HOST="$2"
            shift 2
            ;;
        -u|--user)
            REMOTE_USER="$2"
            shift 2
            ;;
        -p|--path)
            REMOTE_PATH="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 -h <host> -u <user> [-p <remote_path>]"
            echo ""
            echo "Options:"
            echo "  -h, --host     Remote host IP or hostname (required)"
            echo "  -u, --user     Remote user for SSH (required)"
            echo "  -p, --path     Remote path for deployment (default: /tmp/ngrok-migrated)"
            echo "  --help         Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Validate required arguments
if [ -z "$REMOTE_HOST" ] || [ -z "$REMOTE_USER" ]; then
    echo -e "${RED}Error: Both host and user are required${NC}"
    echo "Usage: $0 -h <host> -u <user> [-p <remote_path>]"
    exit 1
fi

echo -e "${YELLOW}Deployment Configuration:${NC}"
echo "  Remote Host: $REMOTE_HOST"
echo "  Remote User: $REMOTE_USER"
echo "  Remote Path: $REMOTE_PATH"
echo ""

# Step 1: Build locally
echo -e "${YELLOW}Step 1: Building binaries locally...${NC}"
make clean
make release-server
make release-client

if [ ! -f "bin/ngrokd" ] || [ ! -f "bin/ngrok" ]; then
    echo -e "${RED}Error: Build failed. Binaries not found.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Binaries built successfully${NC}"

# Step 2: Create deployment package
echo -e "${YELLOW}Step 2: Creating deployment package...${NC}"
TEMP_DIR=$(mktemp -d)
PACKAGE_NAME="ngrok-go124-$(date +%Y%m%d-%H%M%S).tar.gz"

# Copy necessary files
cp -r go.mod go.sum Makefile src bin "$TEMP_DIR/"
cd "$TEMP_DIR"
tar -czf "/tmp/$PACKAGE_NAME" *
cd - > /dev/null
rm -rf "$TEMP_DIR"

echo -e "${GREEN}✓ Package created: /tmp/$PACKAGE_NAME${NC}"

# Step 3: Transfer to remote server
echo -e "${YELLOW}Step 3: Transferring to remote server...${NC}"
scp "/tmp/$PACKAGE_NAME" "$REMOTE_USER@$REMOTE_HOST:/tmp/"

# Step 4: Deploy on remote server
echo -e "${YELLOW}Step 4: Deploying on remote server...${NC}"
ssh "$REMOTE_USER@$REMOTE_HOST" << EOF
set -e

# Extract package
echo "Extracting package..."
mkdir -p "$REMOTE_PATH"
cd "$REMOTE_PATH"
tar -xzf "/tmp/$PACKAGE_NAME"

# Create directories
echo "Creating directories..."
sudo mkdir -p /opt/ngrok/{bin,config,logs,data,certs}
sudo mkdir -p /var/log/ngrok
sudo mkdir -p /etc/ngrok

# Create ngrok user if it doesn't exist
if ! id -u ngrok >/dev/null 2>&1; then
    echo "Creating ngrok user..."
    sudo useradd -r -s /bin/false -d /opt/ngrok ngrok
fi

# Install binaries
echo "Installing binaries..."
sudo cp bin/ngrokd /opt/ngrok/bin/
sudo cp bin/ngrok /opt/ngrok/bin/
sudo chmod +x /opt/ngrok/bin/ngrok*
sudo chown -R ngrok:ngrok /opt/ngrok /var/log/ngrok /etc/ngrok

echo "Deployment completed!"
EOF

# Step 5: Create systemd service
echo -e "${YELLOW}Step 5: Creating systemd service...${NC}"
ssh "$REMOTE_USER@$REMOTE_HOST" << 'EOF'
sudo tee /etc/systemd/system/ngrokd.service > /dev/null << 'SERVICE'
[Unit]
Description=ngrok server (Go 1.24 version)
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=ngrok
Group=ngrok
ExecStart=/opt/ngrok/bin/ngrokd -config=/etc/ngrok/ngrokd.yml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ngrokd

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/ngrok /opt/ngrok/data
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictAddressFamilies=AF_INET AF_INET6
RestrictNamespaces=true
RestrictRealtime=true
RestrictSUIDSGID=true
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM

# Resource limits
LimitNOFILE=65535
LimitNPROC=65535

[Install]
WantedBy=multi-user.target
SERVICE

sudo systemctl daemon-reload
echo "Systemd service created"
EOF

# Cleanup
rm -f "/tmp/$PACKAGE_NAME"

echo ""
echo -e "${GREEN}Deployment completed successfully!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. SSH to $REMOTE_HOST"
echo "2. Create configuration file: sudo nano /etc/ngrok/ngrokd.yml"
echo "3. Generate TLS certificates (see deployment guide)"
echo "4. Start the service: sudo systemctl start ngrokd"
echo "5. Enable auto-start: sudo systemctl enable ngrokd"
echo ""
echo "Example configuration (/etc/ngrok/ngrokd.yml):"
echo "---"
echo "server_addr: \"0.0.0.0:4443\""
echo "http_addr: \"0.0.0.0:80\""
echo "https_addr: \"0.0.0.0:443\""
echo "domain: \"tunnel.yourdomain.com\""
echo "tls_cert: \"/opt/ngrok/certs/server.crt\""
echo "tls_key: \"/opt/ngrok/certs/server.key\""
echo "log_to: \"/var/log/ngrok/ngrokd.log\""
echo "log_level: \"INFO\""