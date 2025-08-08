#!/bin/bash
# Start ngrok HTTPS tunnel to your server

echo "Starting ngrok HTTPS tunnel..."
echo "Server: tunels.tech"
echo "Local port: 8000"
echo ""

# Start the tunnel with HTTPS protocol
./bin/ngrok \
    -config=ngrok.yml \
    -subdomain=${1:-myapp} \
    -proto=https \
    -log=stdout \
    -log-level=DEBUG \
    localhost:8000