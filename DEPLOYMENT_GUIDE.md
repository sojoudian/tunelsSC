# ngrok Deployment Guide for Debian 12 on Azure

> **IMPORTANT**: This guide is for deploying the **Go 1.24 migrated version** of ngrok, NOT the original Go 1.3 version. 
> Make sure you're using the code from the `go-1.24-migration/migrated` directory.

## Table of Contents
1. [Environment Setup](#environment-setup)
2. [Application Deployment](#application-deployment)
3. [Security & Authentication](#security--authentication)
4. [Service Management](#service-management)
5. [Testing & Verification](#testing--verification)
6. [Additional Considerations](#additional-considerations)

---

## Environment Setup

### System Requirements

**Minimum VM Requirements:**
- **VM Size**: B2s (2 vCPUs, 4 GB RAM) or higher
- **Storage**: 30 GB SSD (Premium SSD recommended)
- **OS**: Debian 12 (Bookworm) - Latest from Azure Marketplace
- **Network**: Standard public IP with NSG configured

### Initial System Setup

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install essential packages
sudo apt install -y \
    build-essential \
    git \
    curl \
    wget \
    htop \
    net-tools \
    ufw \
    certbot \
    nginx \
    supervisor \
    rsyslog \
    logrotate

# Install security updates automatically
sudo apt install -y unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades
```

### Go Installation

```bash
# Set Go version
GO_VERSION="1.24"

# Download and install Go
cd /tmp
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# Set up Go environment
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
echo 'export GOPATH=$HOME/go' | tee -a ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' | tee -a ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### System Dependencies

```bash
# Install additional dependencies for ngrok
sudo apt install -y \
    libssl-dev \
    ca-certificates \
    gnupg \
    lsb-release

# Set up system limits for high connection count
sudo tee -a /etc/security/limits.conf <<EOF
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
EOF

# Configure sysctl for network optimization
sudo tee /etc/sysctl.d/99-ngrok.conf <<EOF
# Network optimizations for ngrok
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_max_tw_buckets = 500000
net.ipv4.tcp_tw_reuse = 1
net.ipv4.ip_local_port_range = 1024 65535
EOF

sudo sysctl -p /etc/sysctl.d/99-ngrok.conf
```

---

## Application Deployment

### Directory Structure

```bash
# Create application directories
sudo mkdir -p /opt/ngrok/{bin,config,logs,data,certs}
sudo mkdir -p /var/log/ngrok
sudo mkdir -p /etc/ngrok

# Create ngrok user
sudo useradd -r -s /bin/false -d /opt/ngrok ngrok
sudo chown -R ngrok:ngrok /opt/ngrok /var/log/ngrok /etc/ngrok
```

### Building the Server Component

```bash
# Copy the migrated ngrok Go 1.24 code to your server
# Option 1: SCP from your local machine
scp -r /path/to/ngrok/go-1.24-migration/migrated user@your-server:/tmp/ngrok-migrated

# Option 2: Create a tarball and transfer
cd /path/to/ngrok/go-1.24-migration
tar -czf ngrok-migrated.tar.gz migrated/
scp ngrok-migrated.tar.gz user@your-server:/tmp/

# On the server, extract and build
cd /tmp
tar -xzf ngrok-migrated.tar.gz  # if using Option 2
cd ngrok-migrated  # or cd migrated if using Option 1

# Build server binary
make clean
make release-server

# Install server binary
sudo cp bin/ngrokd /opt/ngrok/bin/
sudo chmod +x /opt/ngrok/bin/ngrokd
sudo chown ngrok:ngrok /opt/ngrok/bin/ngrokd
```

### Building the Client Component

```bash
# In the same migrated directory
# Build client binary
make release-client

# Install client binary (for testing)
sudo cp bin/ngrok /opt/ngrok/bin/
sudo chmod +x /opt/ngrok/bin/ngrok
sudo chown ngrok:ngrok /opt/ngrok/bin/ngrok

# Create client package for distribution
tar -czf ngrok-client-linux-amd64.tar.gz -C bin ngrok

# Note: The client binary should be distributed to end users
# who will connect to your ngrok server
```

### Server Configuration

```bash
# Create server configuration
sudo tee /etc/ngrok/ngrokd.yml <<EOF
# ngrokd configuration
server_addr: "0.0.0.0:4443"
http_addr: "0.0.0.0:80"
https_addr: "0.0.0.0:443"
domain: "tunnel.yourdomain.com"
tls_cert: "/opt/ngrok/certs/server.crt"
tls_key: "/opt/ngrok/certs/server.key"
log_to: "/var/log/ngrok/ngrokd.log"
log_level: "INFO"
EOF

sudo chown ngrok:ngrok /etc/ngrok/ngrokd.yml
sudo chmod 640 /etc/ngrok/ngrokd.yml
```

### TLS Certificate Setup

```bash
# Option 1: Generate self-signed certificates (for testing)
cd /opt/ngrok/certs
sudo openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
    -days 365 -nodes -subj "/CN=tunnel.yourdomain.com"

# Option 2: Use Let's Encrypt (recommended for production)
sudo certbot certonly --standalone -d tunnel.yourdomain.com
sudo ln -s /etc/letsencrypt/live/tunnel.yourdomain.com/fullchain.pem /opt/ngrok/certs/server.crt
sudo ln -s /etc/letsencrypt/live/tunnel.yourdomain.com/privkey.pem /opt/ngrok/certs/server.key

# Set proper permissions
sudo chown -R ngrok:ngrok /opt/ngrok/certs
sudo chmod 600 /opt/ngrok/certs/server.key
sudo chmod 644 /opt/ngrok/certs/server.crt
```

---

## Security & Authentication

### Token Generation

```bash
# Generate authentication tokens
#!/bin/bash
# save as /opt/ngrok/bin/generate-token.sh

generate_token() {
    openssl rand -hex 32
}

# Generate tokens for clients
TOKEN=$(generate_token)
echo "Client Token: $TOKEN"

# Store tokens securely
sudo tee -a /etc/ngrok/auth_tokens.txt <<EOF
$(date +%Y-%m-%d_%H:%M:%S) $TOKEN client_description
EOF

sudo chmod 600 /etc/ngrok/auth_tokens.txt
sudo chown ngrok:ngrok /etc/ngrok/auth_tokens.txt
```

### Firewall Configuration

```bash
# Configure UFW firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (adjust port if needed)
sudo ufw allow 22/tcp

# Allow ngrok ports
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 4443/tcp  # Control port

# Enable firewall
sudo ufw --force enable
sudo ufw status verbose
```

### Azure Network Security Group (NSG)

```bash
# Create NSG rules via Azure CLI
az network nsg rule create \
    --resource-group YOUR_RG \
    --nsg-name YOUR_NSG \
    --name AllowNgrokHTTP \
    --priority 100 \
    --source-address-prefixes "*" \
    --destination-port-ranges 80 \
    --protocol Tcp \
    --access Allow

az network nsg rule create \
    --resource-group YOUR_RG \
    --nsg-name YOUR_NSG \
    --name AllowNgrokHTTPS \
    --priority 101 \
    --source-address-prefixes "*" \
    --destination-port-ranges 443 \
    --protocol Tcp \
    --access Allow

az network nsg rule create \
    --resource-group YOUR_RG \
    --nsg-name YOUR_NSG \
    --name AllowNgrokControl \
    --priority 102 \
    --source-address-prefixes "*" \
    --destination-port-ranges 4443 \
    --protocol Tcp \
    --access Allow
```

### Security Best Practices

```bash
# 1. Disable root SSH login
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# 2. Set up fail2ban
sudo apt install -y fail2ban
sudo tee /etc/fail2ban/jail.local <<EOF
[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[ngrok]
enabled = true
port = 4443
filter = ngrok
logpath = /var/log/ngrok/ngrokd.log
maxretry = 5
bantime = 3600
EOF

# 3. Create fail2ban filter for ngrok
sudo tee /etc/fail2ban/filter.d/ngrok.conf <<EOF
[Definition]
failregex = ^.*Failed authentication from <HOST>.*$
            ^.*Invalid token from <HOST>.*$
ignoreregex =
EOF

sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# 4. Enable audit logging
sudo apt install -y auditd
sudo systemctl enable auditd
sudo systemctl start auditd
```

---

## Service Management

### Systemd Service Configuration

```bash
# Create systemd service for ngrokd
sudo tee /etc/systemd/system/ngrokd.service <<EOF
[Unit]
Description=ngrok server
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
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable ngrokd
sudo systemctl start ngrokd
sudo systemctl status ngrokd
```

### Process Monitoring with Monit

```bash
# Install and configure monit
sudo apt install -y monit

sudo tee /etc/monit/conf.d/ngrokd <<EOF
check process ngrokd with pidfile /var/run/ngrokd.pid
    start program = "/bin/systemctl start ngrokd"
    stop program = "/bin/systemctl stop ngrokd"
    if cpu > 80% for 5 cycles then alert
    if memory > 75% for 5 cycles then alert
    if failed host 127.0.0.1 port 4443 protocol https then restart
    if 3 restarts within 5 cycles then timeout
EOF

sudo systemctl enable monit
sudo systemctl start monit
```

### Log Management

```bash
# Configure rsyslog for ngrok
sudo tee /etc/rsyslog.d/50-ngrok.conf <<EOF
if \$programname == 'ngrokd' then /var/log/ngrok/ngrokd.log
& stop
EOF

# Configure logrotate
sudo tee /etc/logrotate.d/ngrok <<EOF
/var/log/ngrok/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0640 ngrok ngrok
    sharedscripts
    postrotate
        systemctl reload ngrokd > /dev/null 2>&1 || true
    endscript
}
EOF

sudo systemctl restart rsyslog
```

### Monitoring Script

```bash
#!/bin/bash
# Save as /opt/ngrok/bin/health-check.sh

check_service() {
    if systemctl is-active --quiet ngrokd; then
        echo "✓ ngrokd service is running"
    else
        echo "✗ ngrokd service is down"
        exit 1
    fi
}

check_ports() {
    for port in 80 443 4443; do
        if ss -tlnp | grep -q ":$port "; then
            echo "✓ Port $port is listening"
        else
            echo "✗ Port $port is not listening"
            exit 1
        fi
    done
}

check_disk_space() {
    usage=$(df -h /opt/ngrok | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ $usage -lt 90 ]; then
        echo "✓ Disk usage: ${usage}%"
    else
        echo "✗ Disk usage critical: ${usage}%"
        exit 1
    fi
}

check_service
check_ports
check_disk_space
```

---

## Testing & Verification

### Basic Connectivity Tests

```bash
# 1. Check service status
sudo systemctl status ngrokd
sudo journalctl -u ngrokd -f

# 2. Test local connectivity
curl -k https://localhost:4443/

# 3. Test from remote client
# On client machine:
cat > ~/.ngrok <<EOF
server_addr: "tunnel.yourdomain.com:4443"
trust_host_root_certs: false
auth_token: "YOUR_GENERATED_TOKEN"
EOF

./ngrok -config ~/.ngrok -log=stdout http 8080

# 4. DNS verification
nslookup tunnel.yourdomain.com
dig tunnel.yourdomain.com
```

### Performance Testing

```bash
# Install Apache Bench
sudo apt install -y apache2-utils

# Basic load test
ab -n 1000 -c 50 https://tunnel.yourdomain.com/

# Monitor during test
htop
sudo iotop
sudo nethogs
```

### Health Check Endpoint

```bash
# Add health check to ngrok server (if not implemented)
# Test with:
curl -k https://tunnel.yourdomain.com/health

# Set up external monitoring (e.g., Azure Monitor)
az monitor metric alert create \
    --name ngrok-health \
    --resource-group YOUR_RG \
    --scopes /subscriptions/SUB_ID/resourceGroups/RG/providers/Microsoft.Compute/virtualMachines/VM_NAME \
    --condition "avg Percentage CPU > 80" \
    --window-size 5m \
    --evaluation-frequency 1m
```

---

## Additional Considerations

### Performance Optimization

```bash
# 1. Enable TCP BBR congestion control
echo "net.core.default_qdisc=fq" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_congestion_control=bbr" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 2. Optimize Go runtime
# Add to systemd service:
Environment="GOGC=100"
Environment="GOMEMLIMIT=3GiB"
Environment="GOMAXPROCS=2"

# 3. Use Azure Accelerated Networking
az network nic update \
    --name YOUR_NIC \
    --resource-group YOUR_RG \
    --accelerated-networking true
```

### Backup Procedures

```bash
#!/bin/bash
# Save as /opt/ngrok/bin/backup.sh

BACKUP_DIR="/opt/ngrok/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup configuration
tar -czf $BACKUP_DIR/ngrok-config-$DATE.tar.gz \
    /etc/ngrok \
    /opt/ngrok/config \
    /opt/ngrok/certs

# Backup data
tar -czf $BACKUP_DIR/ngrok-data-$DATE.tar.gz \
    /opt/ngrok/data

# Keep only last 7 days of backups
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

# Sync to Azure Blob Storage
az storage blob upload-batch \
    --destination ngrok-backups \
    --source $BACKUP_DIR \
    --pattern "ngrok-*-$DATE.tar.gz"
```

### Maintenance Procedures

```bash
# 1. Certificate renewal (Let's Encrypt)
sudo tee /etc/cron.d/certbot-renewal <<EOF
0 2 * * * root certbot renew --quiet --post-hook "systemctl reload ngrokd"
EOF

# 2. System updates
sudo tee /usr/local/bin/maintenance.sh <<EOF
#!/bin/bash
# Maintenance window script

# Notify users
echo "Maintenance starting in 5 minutes" | wall

sleep 300

# Stop service
systemctl stop ngrokd

# Apply updates
apt update && apt upgrade -y

# Clean logs older than 30 days
find /var/log/ngrok -name "*.log" -mtime +30 -delete

# Start service
systemctl start ngrokd

# Verify
/opt/ngrok/bin/health-check.sh
EOF

sudo chmod +x /usr/local/bin/maintenance.sh
```

### Azure-Specific Configurations

```bash
# 1. Azure VM Auto-shutdown
az vm auto-shutdown \
    --resource-group YOUR_RG \
    --name YOUR_VM \
    --time 2300

# 2. Azure Backup
az backup protection enable-for-vm \
    --resource-group YOUR_RG \
    --vault-name YOUR_VAULT \
    --vm YOUR_VM \
    --policy-name DefaultPolicy

# 3. Azure Monitor Agent
wget https://github.com/microsoft/OMS-Agent-for-Linux/releases/download/OMSAgent_v1.14.19-0/omsagent-1.14.19-0.universal.x64.sh
sudo sh omsagent-1.14.19-0.universal.x64.sh --install

# 4. Enable Azure Diagnostics
az vm diagnostics set \
    --resource-group YOUR_RG \
    --vm-name YOUR_VM \
    --settings @diagnostic-config.json
```

### Troubleshooting Guide

```bash
# Common issues and solutions:

# 1. Service won't start
sudo journalctl -xe -u ngrokd
sudo systemctl status ngrokd.service

# 2. Port already in use
sudo lsof -i :4443
sudo netstat -tlnp | grep 4443

# 3. Certificate issues
openssl x509 -in /opt/ngrok/certs/server.crt -text -noout
openssl verify /opt/ngrok/certs/server.crt

# 4. Permission issues
sudo -u ngrok ls -la /opt/ngrok/
sudo chown -R ngrok:ngrok /opt/ngrok /var/log/ngrok

# 5. High CPU/Memory usage
top -u ngrok
ps aux | grep ngrok
sudo strace -p $(pgrep ngrokd)

# 6. Connection issues
sudo tcpdump -i any port 4443 -w ngrok.pcap
curl -vvv -k https://localhost:4443/
```

### Production Checklist

- [ ] VM properly sized for expected load
- [ ] Firewall rules configured and tested
- [ ] TLS certificates installed and valid
- [ ] Service running and monitored
- [ ] Backups configured and tested
- [ ] Logs rotating properly
- [ ] Health checks passing
- [ ] Azure monitoring enabled
- [ ] Security hardening applied
- [ ] Documentation updated
- [ ] Disaster recovery plan in place

---

## Client Configuration Guide

For clients connecting to your ngrok server:

```bash
# Client configuration file (~/.ngrok)
server_addr: "tunnel.yourdomain.com:4443"
trust_host_root_certs: false
auth_token: "YOUR_AUTH_TOKEN"
tunnels:
  webapp:
    proto:
      http: 8080
    subdomain: myapp
  ssh:
    proto:
      tcp: 22
    remote_port: 2222

# Start client
./ngrok start webapp ssh
```

---

This deployment guide provides a production-ready setup for ngrok on Azure Debian 12. Adjust configurations based on your specific requirements and load expectations.