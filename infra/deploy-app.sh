#!/bin/bash

# Deploy PulseExpends application to Huawei Cloud ECS
# This script copies the local application and deploys it

set -e

KEY_FILE="/opt/accessKeys/pulse-expends-key.pem"
IP="159.138.118.60"
APP_SOURCE="/root/PulseExpends"
APP_DEST="/opt/pulse-expends"

echo "=== Deploying PulseExpends Application ==="
echo "Target: $IP"
echo "Source: $APP_SOURCE"
echo "Destination: $APP_DEST"
echo ""

# Check prerequisites
if [ ! -f "$KEY_FILE" ]; then
    echo "❌ ERROR: SSH key file '$KEY_FILE' not found!"
    echo "   Make sure you've imported the key to Huawei Cloud and bound it to the instance."
    exit 1
fi

if [ ! -d "$APP_SOURCE" ]; then
    echo "❌ ERROR: Application source directory '$APP_SOURCE' not found!"
    exit 1
fi

# Test SSH connection first
echo "🔐 Testing SSH connection..."
ssh -i "$KEY_FILE" -o ConnectTimeout=10 -o StrictHostKeyChecking=no -o BatchMode=yes "root@$IP" "echo 'SSH connection successful'" >/dev/null 2>&1

if [ $? -ne 0 ]; then
    echo "❌ SSH connection failed!"
    echo ""
    echo "📋 Troubleshooting steps:"
    echo "1. Go to Huawei Cloud Console → ECS → Instances"
    echo "2. Find instance: acc9edeb-1cd4-4181-8904-881135933ac5"
    echo "3. Ensure key pair 'pulse-expends-key' is bound to the instance"
    echo "4. Try 'Remote Login' (VNC) to check instance status"
    echo "5. The instance might need to be restarted after key pair change"
    echo ""
    echo "🔑 Public key to import (if needed):"
    echo "---"
    ssh-keygen -y -f "$KEY_FILE"
    echo "---"
    exit 1
fi

echo "✅ SSH connection successful"

# Create deployment directory on remote server
echo "📁 Creating deployment directory..."
ssh -i "$KEY_FILE" -o StrictHostKeyChecking=no "root@$IP" "mkdir -p $APP_DEST && rm -rf $APP_DEST/*"

# Copy application files
echo "📤 Copying application files..."
scp -i "$KEY_FILE" -o StrictHostKeyChecking=no -r "$APP_SOURCE/" "root@$IP:$APP_DEST/"

# Create fixed docker-compose.yml (without GitHub dependency)
echo "🐳 Creating fixed docker-compose.yml..."
cat > /tmp/docker-compose-fixed.yml << 'EOF'
version: '3.8'

services:
  # MCP Server (Go)
  mcp-server:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=dev
      - OBS_ENDPOINT=https://obs.la-south-2.myhuaweicloud.com
      - OBS_ACCESS_KEY=${OBS_ACCESS_KEY}
      - OBS_SECRET_KEY=${OBS_SECRET_KEY}
      - OBS_BUCKET_NAME=pulse-expends-data-dev-pulse-afd05cc1
      - JWT_SECRET=pulse-expends-jwt-secret-prod-2024-change-me
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    restart: unless-stopped
    networks:
      - pulseexpends-network

  # Python PDF Parser Service
  pdf-parser:
    build:
      context: ./python/pdf-parser
      dockerfile: Dockerfile
    ports:
      - "8000:8000"
    environment:
      - ENVIRONMENT=dev
      - OBS_ENDPOINT=https://obs.la-south-2.myhuaweicloud.com
      - OBS_ACCESS_KEY=${OBS_ACCESS_KEY}
      - OBS_SECRET_KEY=${OBS_SECRET_KEY}
      - OBS_BUCKET_NAME=pulse-expends-data-dev-pulse-afd05cc1
      - PYTHON_SERVICE_URL=http://localhost:8000
    volumes:
      - ./python/pdf-parser/logs:/app/logs
      - ./python/pdf-parser/uploads:/app/uploads
    restart: unless-stopped
    networks:
      - pulseexpends-network

  # Nginx Reverse Proxy
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/logs:/var/log/nginx
    restart: unless-stopped
    networks:
      - pulseexpends-network
    depends_on:
      - mcp-server
      - pdf-parser

networks:
  pulseexpends-network:
    driver: bridge
EOF

# Copy fixed docker-compose.yml
scp -i "$KEY_FILE" -o StrictHostKeyChecking=no /tmp/docker-compose-fixed.yml "root@$IP:$APP_DEST/docker-compose.yml"

# Create environment file
echo "⚙️ Creating environment file..."
ssh -i "$KEY_FILE" -o StrictHostKeyChecking=no "root@$IP" "cat > $APP_DEST/.env << 'EOF'
# PulseExpends Environment Variables
ENVIRONMENT=dev
REGION=la-south-2
DOMAIN_NAME=pulseexpends.duckdns.org
ENABLE_SSL=false

# OBS Configuration
OBS_ENDPOINT=https://obs.la-south-2.myhuaweicloud.com
OBS_ACCESS_KEY=HPUAELE34ORKBY58ROT4
OBS_SECRET_KEY=Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439
OBS_BUCKET_NAME=pulse-expends-data-dev-pulse-afd05cc1

# Application Configuration
JWT_SECRET=pulse-expends-jwt-secret-prod-2024-change-me
PYTHON_SERVICE_URL=http://localhost:8000
EOF"

# Create nginx configuration
echo "🌐 Creating nginx configuration..."
ssh -i "$KEY_FILE" -o StrictHostKeyChecking=no "root@$IP" "mkdir -p $APP_DEST/nginx && cat > $APP_DEST/nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '\$remote_addr - \$remote_user [\$time_local] \"\$request\" '
                      '\$status \$body_bytes_sent \"\$http_referer\" '
                      '\"\$http_user_agent\" \"\$http_x_forwarded_for\"';

    access_log  /var/log/nginx/access.log  main;
    error_log   /var/log/nginx/error.log warn;

    sendfile        on;
    keepalive_timeout  65;

    # MCP Server
    server {
        listen 80;
        server_name pulseexpends.duckdns.org localhost;

        location / {
            proxy_pass http://mcp-server:8080;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        location /api/ {
            proxy_pass http://mcp-server:8080;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
    }

    # Python PDF Parser Service
    server {
        listen 8000;
        server_name pulseexpends.duckdns.org localhost;

        location / {
            proxy_pass http://pdf-parser:8000;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
    }
}
EOF"

# Install Docker if not present
echo "🐳 Installing Docker..."
ssh -i "$KEY_FILE" -o StrictHostKeyChecking=no "root@$IP" "which docker >/dev/null 2>&1 || (apt-get update && apt-get install -y docker.io docker-compose)"

# Build and start services
echo "🚀 Building and starting services..."
ssh -i "$KEY_FILE" -o StrictHostKeyChecking=no "root@$IP" "cd $APP_DEST && docker-compose up -d --build"

# Check service status
echo "🔍 Checking service status..."
ssh -i "$KEY_FILE" -o StrictHostKeyChecking=no "root@$IP" "cd $APP_DEST && docker-compose ps"

# Test endpoints
echo "🌐 Testing application endpoints..."
echo "Testing port 8080 (MCP Server)..."
if curl -s --connect-timeout 5 http://$IP:8080 >/dev/null 2>&1; then
    echo "✅ MCP Server (8080) is accessible"
else
    echo "❌ MCP Server (8080) is not accessible"
fi

echo "Testing port 8000 (PDF Parser)..."
if curl -s --connect-timeout 5 http://$IP:8000 >/dev/null 2>&1; then
    echo "✅ PDF Parser (8000) is accessible"
else
    echo "❌ PDF Parser (8000) is not accessible"
fi

echo "Testing port 80 (Nginx)..."
if curl -s --connect-timeout 5 http://$IP >/dev/null 2>&1; then
    echo "✅ Nginx (80) is accessible"
else
    echo "❌ Nginx (80) is not accessible"
fi

echo ""
echo "🎉 Deployment complete!"
echo ""
echo "📋 Application URLs:"
echo "   - MCP Server: http://$IP:8080"
echo "   - PDF Parser: http://$IP:8000"
echo "   - Main App: http://$IP"
echo "   - Domain: http://pulseexpends.duckdns.org"
echo ""
echo "🔧 To check logs:"
echo "   ssh -i $KEY_FILE root@$IP"
echo "   cd $APP_DEST"
echo "   docker-compose logs -f"
echo ""
echo "🔄 To restart services:"
echo "   cd $APP_DEST && docker-compose restart"
echo ""
echo "🗑️ To stop services:"
echo "   cd $APP_DEST && docker-compose down"