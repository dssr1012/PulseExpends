#!/bin/bash

# PulseExpends Authentication System Deployment Script
# Deploys the complete authentication system with Google OAuth and family groups

set -e

echo "=============================================="
echo "  PulseExpends - Authentication Deployment    "
echo "=============================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    print_warning "Running without root privileges. Some commands may require sudo."
    echo ""
fi

# Step 1: Check prerequisites
echo "🔍 Checking prerequisites..."
echo ""

# Check Go
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21+"
    echo "Install Go: https://golang.org/dl/"
    exit 1
else
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Go $GO_VERSION installed"
fi

# Check PostgreSQL
if ! command -v psql &> /dev/null; then
    print_error "PostgreSQL is not installed"
    echo "Install PostgreSQL:"
    echo "  Ubuntu/Debian: sudo apt install postgresql postgresql-contrib"
    echo "  CentOS/RHEL: sudo yum install postgresql-server postgresql-contrib"
    exit 1
else
    PG_VERSION=$(psql --version | awk '{print $3}')
    print_status "PostgreSQL $PG_VERSION installed"
fi

# Check if PostgreSQL is running
if ! sudo systemctl is-active --quiet postgresql; then
    print_warning "PostgreSQL is not running. Starting..."
    sudo systemctl start postgresql
    sudo systemctl enable postgresql
    print_status "PostgreSQL started and enabled"
fi

# Step 2: Setup database
echo ""
echo "🗄️  Setting up database..."
echo ""

# Create database and user if they don't exist
sudo -u postgres psql << EOF
DO \$\$
BEGIN
    -- Create database if not exists
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pulseexpends') THEN
        CREATE DATABASE pulseexpends;
        RAISE NOTICE 'Database pulseexpends created';
    ELSE
        RAISE NOTICE 'Database pulseexpends already exists';
    END IF;
    
    -- Create user if not exists
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'pulseexpends') THEN
        CREATE USER pulseexpends WITH PASSWORD 'pulseexpends_password';
        RAISE NOTICE 'User pulseexpends created';
    ELSE
        RAISE NOTICE 'User pulseexpends already exists';
    END IF;
    
    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE pulseexpends TO pulseexpends;
    
    -- Create extensions
    \c pulseexpends
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pgcrypto";
    
    -- Grant schema privileges
    GRANT ALL ON SCHEMA public TO pulseexpends;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pulseexpends;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO pulseexpends;
END
\$\$;
EOF

if [ $? -eq 0 ]; then
    print_status "Database setup completed"
else
    print_error "Failed to setup database"
    exit 1
fi

# Step 3: Install Go dependencies
echo ""
echo "📦 Installing Go dependencies..."
echo ""

cd backend/auth
go mod download
cd ../..

print_status "Go dependencies installed"

# Step 4: Build authentication server
echo ""
echo "🔨 Building authentication server..."
echo ""

cd backend/auth
go build -o auth-server main.go
cd ../..

if [ $? -eq 0 ]; then
    print_status "Authentication server built successfully"
else
    print_error "Failed to build authentication server"
    exit 1
fi

# Step 5: Setup environment
echo ""
echo "🔧 Setting up environment..."
echo ""

# Create auth directory in /opt
sudo mkdir -p /opt/PulseExpends/backend/auth
sudo cp -r backend/auth/* /opt/PulseExpends/backend/auth/

# Create .env file if it doesn't exist
ENV_FILE="/opt/PulseExpends/backend/auth/.env"
if [ ! -f "$ENV_FILE" ]; then
    print_warning ".env file not found. Creating from template..."
    sudo cp backend/auth/.env.example "$ENV_FILE"
    
    # Generate secure secrets
    sudo sh -c "echo 'JWT_SECRET=$(openssl rand -hex 32)' >> $ENV_FILE"
    sudo sh -c "echo 'COOKIE_SECRET=$(openssl rand -hex 32)' >> $ENV_FILE"
    sudo sh -c "echo 'SESSION_SECRET=$(openssl rand -hex 32)' >> $ENV_FILE"
    
    print_status ".env file created at $ENV_FILE"
    print_warning "Please update Google OAuth credentials in $ENV_FILE"
else
    print_status ".env file already exists"
fi

# Set permissions
sudo chown -R www-data:www-data /opt/PulseExpends/backend/auth
sudo chmod 750 /opt/PulseExpends/backend/auth
sudo chmod 640 "$ENV_FILE"

print_status "Environment setup completed"

# Step 6: Setup systemd service
echo ""
echo "⚙️  Setting up systemd service..."
echo ""

SERVICE_FILE="/etc/systemd/system/pulseexpends-auth-server.service"
sudo cp backend/auth/systemd/pulseexpends-auth-server.service "$SERVICE_FILE"

# Reload systemd
sudo systemctl daemon-reload
sudo systemctl enable pulseexpends-auth-server

print_status "Systemd service configured"

# Step 7: Configure Nginx
echo ""
echo "🌐 Configuring Nginx..."
echo ""

NGINX_CONF="/etc/nginx/sites-available/pulseexpends-auth"
if [ ! -f "$NGINX_CONF" ]; then
    sudo tee "$NGINX_CONF" > /dev/null << EOF
# PulseExpends Authentication Server
server {
    listen 80;
    server_name api.pulseexpends.duckdns.org;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # CORS headers
    add_header Access-Control-Allow-Origin "http://pulseexpends.duckdns.org" always;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
    add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With" always;
    add_header Access-Control-Allow-Credentials "true" always;
    
    # Handle preflight requests
    if (\$request_method = 'OPTIONS') {
        add_header Access-Control-Allow-Origin "http://pulseexpends.duckdns.org";
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With";
        add_header Access-Control-Allow-Credentials "true";
        add_header Content-Length 0;
        add_header Content-Type text/plain;
        return 204;
    }
    
    # Proxy to auth server
    location / {
        proxy_pass http://localhost:8082;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffer sizes
        proxy_buffer_size 128k;
        proxy_buffers 4 256k;
        proxy_busy_buffers_size 256k;
    }
    
    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8082/health;
        access_log off;
    }
    
    # Logging
    access_log /var/log/nginx/pulseexpends-auth.access.log;
    error_log /var/log/nginx/pulseexpends-auth.error.log;
}
EOF
    
    # Enable site
    sudo ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/
    
    # Test and reload Nginx
    sudo nginx -t && sudo systemctl reload nginx
    
    print_status "Nginx configuration created"
else
    print_status "Nginx configuration already exists"
fi

# Step 8: Update frontend configuration
echo ""
echo "🎨 Updating frontend configuration..."
echo ""

# Update auth.js with correct API URL
FRONTEND_AUTH_JS="frontend/src/auth.js"
if [ -f "$FRONTEND_AUTH_JS" ]; then
    # Check if already configured
    if ! grep -q "api.pulseexpends.duckdns.org" "$FRONTEND_AUTH_JS"; then
        sed -i 's|http://localhost:8082/api|http://api.pulseexpends.duckdns.org/api|g' "$FRONTEND_AUTH_JS"
        print_status "Frontend API URL updated for production"
    else
        print_status "Frontend already configured for production"
    fi
fi

# Update app.js with correct API URL
FRONTEND_APP_JS="frontend/src/app.js"
if [ -f "$FRONTEND_APP_JS" ]; then
    # Check if already configured
    if ! grep -q "api.pulseexpends.duckdns.org" "$FRONTEND_APP_JS"; then
        sed -i 's|http://localhost:8082|http://api.pulseexpends.duckdns.org|g' "$FRONTEND_APP_JS"
        print_status "App.js API URL updated for production"
    else
        print_status "App.js already configured for production"
    fi
fi

# Step 9: Start services
echo ""
echo "🚀 Starting services..."
echo ""

# Start auth server
sudo systemctl start pulseexpends-auth-server
sleep 2

# Check if auth server is running
if sudo systemctl is-active --quiet pulseexpends-auth-server; then
    print_status "Authentication server started successfully"
else
    print_error "Failed to start authentication server"
    sudo journalctl -u pulseexpends-auth-server -n 20 --no-pager
    exit 1
fi

# Restart Nginx
sudo systemctl restart nginx
if sudo systemctl is-active --quiet nginx; then
    print_status "Nginx restarted successfully"
else
    print_error "Failed to restart Nginx"
    sudo journalctl -u nginx -n 20 --no-pager
    exit 1
fi

# Step 10: Verify deployment
echo ""
echo "🔍 Verifying deployment..."
echo ""

# Test health endpoint
echo "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/health || echo "FAILED")
if [ "$HEALTH_RESPONSE" = "200" ]; then
    print_status "Health check passed (HTTP $HEALTH_RESPONSE)"
else
    print_warning "Health check failed (HTTP $HEALTH_RESPONSE)"
    echo "Checking logs..."
    sudo journalctl -u pulseexpends-auth-server -n 10 --no-pager
fi

# Test API endpoint
echo "Testing API endpoint..."
API_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/api/health || echo "FAILED")
if [ "$API_RESPONSE" = "200" ]; then
    print_status "API endpoint working (HTTP $API_RESPONSE)"
else
    print_warning "API endpoint check failed (HTTP $API_RESPONSE)"
fi

# Test Nginx proxy
echo "Testing Nginx proxy..."
NGINX_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://api.pulseexpends.duckdns.org/health || echo "FAILED")
if [ "$NGINX_RESPONSE" = "200" ]; then
    print_status "Nginx proxy working (HTTP $NGINX_RESPONSE)"
else
    print_warning "Nginx proxy check failed (HTTP $NGINX_RESPONSE)"
    echo "Make sure DNS is configured correctly for api.pulseexpends.duckdns.org"
fi

# Step 11: Create admin user (optional)
echo ""
echo "👤 Creating admin user (optional)..."
echo ""

read -p "Do you want to create an admin user? (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "Enter admin email: " ADMIN_EMAIL
    read -p "Enter admin password: " -s ADMIN_PASSWORD
    echo ""
    read -p "Enter admin full name: " ADMIN_NAME
    
    # Create admin user via API
    curl -X POST http://localhost:8082/api/auth/register \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\",\"fullName\":\"$ADMIN_NAME\",\"username\":\"admin\"}" \
        || echo "Note: Admin user creation might fail if server is still starting up"
    
    print_status "Admin user creation requested"
fi

# Step 12: Final instructions
echo ""
echo "=============================================="
echo "  Deployment Complete!                       "
echo "=============================================="
echo ""
echo "🎉 Authentication system deployed successfully!"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Configure Google OAuth:"
echo "   Run: ./configure-google-oauth.sh"
echo "   Then edit: /opt/PulseExpends/backend/auth/.env"
echo ""
echo "2. Update Google OAuth credentials:"
echo "   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com"
echo "   GOOGLE_CLIENT_SECRET=your-client-secret"
echo "   GOOGLE_REDIRECT_URL=http://api.pulseexpends.duckdns.org/api/auth/google/callback"
echo ""
echo "3. Configure DuckDNS subdomain:"
echo "   Add A record for api.pulseexpends.duckdns.org pointing to your server IP"
echo ""
echo "4. Test the system:"
echo "   Open: http://pulseexpends.duckdns.org/auth/login.html"
echo "   Try: Register with email/password"
echo "   Try: Login with Google (after configuring OAuth)"
echo "   Try: Create a family group"
echo ""
echo "5. Monitor logs:"
echo "   Auth server: sudo journalctl -u pulseexpends-auth-server -f"
echo "   Nginx: sudo tail -f /var/log/nginx/pulseexpends-auth.*.log"
echo ""
echo "6. Backup your configuration:"
echo "   cp /opt/PulseExpends/backend/auth/.env ~/pulseexpends-auth-backup.env"
echo "   pg_dump -U pulseexpends pulseexpends > ~/pulseexpends-db-backup.sql"
echo ""
echo "🔧 Service management:"
echo "   Start: sudo systemctl start pulseexpends-auth-server"
echo "   Stop: sudo systemctl stop pulseexpends-auth-server"
echo "   Restart: sudo systemctl restart pulseexpends-auth-server"
echo "   Status: sudo systemctl status pulseexpends-auth-server"
echo "   Logs: sudo journalctl -u pulseexpends-auth-server -f"
echo ""
echo "🌐 URLs:"
echo "   Frontend: http://pulseexpends.duckdns.org"
echo "   Auth API: http://api.pulseexpends.duckdns.org"
echo "   Health: http://api.pulseexpends.duckdns.org/health"
echo ""
echo "🔐 Security checklist:"
echo "   ☐ Update default passwords in .env"
echo "   ☐ Configure HTTPS with Let's Encrypt"
echo "   ☐ Set up firewall (ufw allow 80,443)"
echo "   ☐ Configure automatic backups"
echo "   ☐ Set up monitoring (Prometheus/Grafana)"
echo ""
echo "📞 Need help?"
echo "   Check logs: sudo journalctl -u pulseexpends-auth-server"
echo "   Check Nginx: sudo nginx -t"
echo "   Test API: curl http://localhost:8082/api/health"
echo "   Database: sudo -u postgres psql -d pulseexpends"
echo ""
echo "=============================================="
echo "  ¡Sistema de autenticación listo!           "
echo "=============================================="