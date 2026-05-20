#!/bin/bash

# Script to configure DuckDNS after Terraform deployment

set -e

echo "DuckDNS Configuration Script"
echo "============================="

# Check if terraform output is available
if ! command -v terraform &> /dev/null; then
    echo "Error: Terraform not found"
    exit 1
fi

# Get the EIP address from Terraform output
echo "Getting EIP address from Terraform..."
EIP_ADDRESS=$(terraform output -raw domain_ip_address 2>/dev/null || true)

if [ -z "$EIP_ADDRESS" ]; then
    echo "Error: Could not get EIP address from Terraform output"
    echo "Please run 'terraform apply' first, then run this script again."
    exit 1
fi

echo "✅ EIP Address: $EIP_ADDRESS"
echo ""

# Ask for DuckDNS token
read -p "Enter your DuckDNS token (get it from https://www.duckdns.org): " DUCKDNS_TOKEN

if [ -z "$DUCKDNS_TOKEN" ]; then
    echo "Error: DuckDNS token is required"
    exit 1
fi

# Domain name
DOMAIN="pulseexpends.duckdns.org"

echo ""
echo "Configuring DuckDNS..."
echo "Domain: $DOMAIN"
echo "IP Address: $EIP_ADDRESS"
echo ""

# Update DuckDNS
echo "Updating DuckDNS record..."
RESPONSE=$(curl -s "https://www.duckdns.org/update?domains=pulseexpends&token=$DUCKDNS_TOKEN&ip=$EIP_ADDRESS")

if [ "$RESPONSE" = "OK" ]; then
    echo "✅ DuckDNS updated successfully!"
    echo ""
    echo "DNS propagation may take 5-10 minutes."
    echo "You can check with: dig +short $DOMAIN"
else
    echo "❌ Failed to update DuckDNS: $RESPONSE"
    echo ""
    echo "Manual update instructions:"
    echo "1. Go to https://www.duckdns.org"
    echo "2. Log in to your account"
    echo "3. Select domain: pulseexpends"
    echo "4. Update IP address to: $EIP_ADDRESS"
    echo "5. Click 'Save'"
fi

echo ""
echo "Application URLs:"
echo "  - Main Application: http://$DOMAIN"
echo "  - MCP Server: http://$DOMAIN:8080"
echo "  - PDF Parser: http://$DOMAIN:8000"
echo ""
echo "SSH Access:"
SSH_CMD=$(terraform output -raw ssh_access_command 2>/dev/null || echo "ssh -i pulse-expends-key.pem root@$EIP_ADDRESS")
echo "  $SSH_CMD"
echo ""
echo "Next steps:"
echo "1. Wait for DNS propagation (5-10 minutes)"
echo "2. Test access: curl http://$DOMAIN"
echo "3. Check application logs: $SSH_CMD 'cd /opt/pulse-expends && docker-compose logs -f'"
echo "4. Configure SSL/TLS when ready (set enable_ssl = true in terraform.tfvars)"