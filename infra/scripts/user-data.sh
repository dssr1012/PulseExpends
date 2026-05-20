#!/bin/bash

# User data script for PulseExpends ECS instances
# This script runs on instance startup

set -e

# Update system
echo "Updating system packages..."
apt-get update -y
apt-get upgrade -y

# Install Docker
echo "Installing Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
rm get-docker.sh

# Install Docker Compose
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create application directory
echo "Creating application directory..."
mkdir -p /opt/pulse-expends
cd /opt/pulse-expends

# Create environment variables file
cat > .env << EOF
# PulseExpends Environment Variables
ENVIRONMENT=${ENVIRONMENT}
REGION=${REGION}
DOMAIN_NAME=${DOMAIN_NAME}
ENABLE_SSL=${ENABLE_SSL}

# OBS Configuration
OBS_ENDPOINT=${OBS_ENDPOINT}
OBS_ACCESS_KEY=${OBS_ACCESS_KEY}
OBS_SECRET_KEY=${OBS_SECRET_KEY}
OBS_BUCKET_NAME=${OBS_BUCKET_NAME}

# Application Configuration
JWT_SECRET=${JWT_SECRET}
PYTHON_SERVICE_URL=${PYTHON_SERVICE_URL}
٫/licenses小说网本实用新型 תש�构成犯罪的和提高<empty ],
EOF

# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  # MCP Server (Go)
  mcp-server:
    image: pulseexpends/mcp-server:latest
    build:
      context: .
      dockerfile: Dockerfile.mcp
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=${ENVIRONMENT}
      - OBS_ENDPOINT=${OBS_ENDPOINT}
      - OBS_ACCESS_KEY=${OBS_ACCESS_KEY}
      - OBS_SECRET_KEY=${OBS_SECRET_KEY}
      - OBS_BUCKET_NAME=${OBS_BUCKET_NAME}
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    restart: unless-stopped
    networks:
      - pulseexpends-network

  # Python PDF Parser Service
  pdf-parser:
    image: pulseexpends/pdf-parser:latest
    build:
      context: ./python/pdf-parser
      dockerfile: Dockerfile
    ports:
      - "8000:8000"
    environment:
      - ENVIRONMENT=${ENVIRONMENT}
      - OBS_ENDPOINT=${OBS_ENDPOINT}
      - OBS_ACCESS_KEY=${OBS_ACCESS_KEY}
      - OBS_SECRET_KEY=${OBS_SECRET_KEY}
      - OBS_BUCKET_NAME=${OBS_BUCKET_NAME}
      - PYTHON_SERVICE_URL=${PYTHON_SERVICE_URL}
    volumes:
      - ./python/pdf-parser/logs:/app/logs
      - ./python/pdf-parser/uploads:/app/uploads
    restart: unless-stopped
    networks:
      - pulseexpends-network

  # Nginx Reverse Proxy (if domain enabled)
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
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

# Create Nginx configuration directory
mkdir -p /opt/pulse-expends/nginx
mkdir -p /opt/pulse-expends/nginx/ssl
mkdir -p /opt/pulse-expends/nginx/logs

# Create Nginx configuration
cat > /opt/pulse-expends/nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;
    error_log   /var/log/nginx/error.log warn;

    sendfile        on;
    keepalive_timeout  65;

    # MCP Server
    server {
        listen 80;
        server_name ${DOMAIN_NAME} localhost;

        location / {
            proxy_pass http://mcp-server:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        location /api/ {
            proxy_pass http://mcp-server:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # Python PDF Parser Service
    server {
        listen 8000;
        server_name ${DOMAIN_NAME} localhost;

        location / {
            proxy_pass http://pdf-parser:8000;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
EOF

# Clone the application repository
echo "Cloning PulseExpends repository..."
git clone https://github.com/dssr1012/PulseExpends.git /opt/pulse-expends/app

# Start services
echo "Starting services..."
cd /opt/pulse-expends
docker-compose up -d

echo "PulseExpends deployment completed!"
echo "Application URLs:"
echo "  - MCP Server: http://localhost:8080"
echo "  - PDF Parser: http://localhost:8000"
if [ -n "$DOMAIN_NAME" ]; then
    echo "  - Domain: http://$DOMAIN_NAME"
fi