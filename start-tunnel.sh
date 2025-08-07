#!/bin/bash
# Start ngrok tunnel to your server

echo "Starting ngrok tunnel..."
echo "Server: tunnel.tunels.tech"
echo "Local port: 8000"
echo ""

# Run ngrok with insecure flag to skip certificate verification
./bin/ngrok -config=ngrok.yml -insecure -subdomain=test -proto=http 8000