#!/bin/bash

# PulseExpends Authentication Database Setup Script
# This script sets up PostgreSQL database for authentication system

set -e

echo "=============================================="
echo "  PulseExpends - Database Setup Script        "
echo "=============================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${YELLOW}⚠️  Warning: Running without root privileges. Some commands may fail.${NC}"
    echo ""
fi

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

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    print_warning "PostgreSQL is not installed. Installing..."
    
    # Detect OS
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
    else
        print_error "Cannot detect OS. Please install PostgreSQL manually."
        exit 1
    fi
    
    case $OS in
        ubuntu|debian)
            sudo apt update
            sudo apt install -y postgresql postgresql-contrib
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            ;;
        centos|rhel|fedora)
            sudo dnf install -y postgresql-server postgresql-contrib
            sudo postgresql-setup --initdb
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            ;;
        *)
            print_error "Unsupported OS: $OS. Please install PostgreSQL manually."
            exit 1
            ;;
    esac
    
    print_status "PostgreSQL installed successfully"
fi

# Check if PostgreSQL is running
if ! sudo systemctl is-active --quiet postgresql; then
    print_warning "PostgreSQL is not running. Starting..."
    sudo systemctl start postgresql
    print_status "PostgreSQL started successfully"
fi

# Create database and user
echo ""
echo "📊 Setting up database..."
echo ""

# Switch to postgres user
sudo -u postgres psql << EOF
-- Create database if not exists
SELECT 'CREATE DATABASE pulseexpends'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pulseexpends')\gexec

-- Create user if not exists
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'pulseexpends') THEN
        CREATE USER pulseexpends WITH PASSWORD 'pulseexpends_password';
    END IF;
END
\$\$;

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

-- Test connection
\q
EOF

if [ $? -eq 0 ]; then
    print_status "Database setup completed successfully"
else
    print_error "Failed to setup database"
    exit 1
fi

# Create .env file for auth service
echo ""
echo "🔧 Creating environment configuration..."
echo ""

AUTH_DIR="backend/auth"
ENV_FILE="$AUTH_DIR/.env"

# Create auth directory if it doesn't exist
mkdir -p "$AUTH_DIR"

# Check if .env file exists
if [ -f "$ENV_FILE" ]; then
    print_warning ".env file already exists. Backing up..."
    cp "$ENV_FILE" "$ENV_FILE.backup.$(date +%Y%m%d_%H%M%S)"
fi

# Create .env file
cat > "$ENV_FILE" << EOF
# Database Configuration
DATABASE_URL=postgresql://pulseexpends:pulseexpends_password@localhost:5432/pulseexpends
DATABASE_MAX_CONNECTIONS=10
DATABASE_IDLE_TIMEOUT=30000
DATABASE_CONNECTION_TIMEOUT=2000

# Server Configuration
PORT=8082
HOST=0.0.0.0
ENVIRONMENT=development

# JWT Configuration
JWT_SECRET=$(openssl rand -hex 32)
JWT_EXPIRY_HOURS=168

# Cookie Configuration
COOKIE_SECRET=$(openssl rand -hex 32)
COOKIE_SECURE=false
COOKIE_DOMAIN=localhost

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080,http://pulseexpends.duckdns.org

# Google OAuth Configuration (Update these with your credentials)
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8082/api/auth/google/callback

# Email Configuration (optional, for password reset)
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USER=your-email@gmail.com
# SMTP_PASSWORD=your-app-password
# SMTP_FROM=noreply@pulseexpends.com

# Frontend URLs
FRONTEND_URL=http://localhost:3000
ADMIN_EMAIL=admin@pulseexpends.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Security
RATE_LIMIT_WINDOW=15
RATE_LIMIT_MAX_REQUESTS=100

# Session Configuration
SESSION_SECRET=$(openssl rand -hex 32)
SESSION_MAX_AGE=604800 # 7 days in seconds

# Password Policy
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBERS=true
PASSWORD_REQUIRE_SPECIAL=false

# Email Verification
EMAIL_VERIFICATION_REQUIRED=false
EMAIL_VERIFICATION_EXPIRY_HOURS=24

# Password Reset
PASSWORD_RESET_EXPIRY_HOURS=1

# Demo Mode
DEMO_MODE=false
DEMO_USER_EMAIL=demo@pulseexpends.com
DEMO_USER_PASSWORD=demo123
EOF

print_status "Environment file created at: $ENV_FILE"

# Install Go dependencies
echo ""
echo "📦 Installing Go dependencies..."
echo ""

cd "$AUTH_DIR"
go mod download
cd - > /dev/null

print_status "Go dependencies installed"

# Build the auth server
echo ""
echo "🔨 Building authentication server..."
echo ""

cd "$AUTH_DIR"
go build -o auth-server main.go
cd - > /dev/null

if [ $? -eq 0 ]; then
    print_status "Authentication server built successfully"
else
    print_error "Failed to build authentication server"
    exit 1
fi

# Create systemd service file
echo ""
echo "⚙️  Creating systemd service..."
echo ""

SERVICE_FILE="/etc/systemd/system/pulseexpends-auth-server.service"

sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=PulseExpends Authentication Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$(pwd)/backend/auth
EnvironmentFile=$(pwd)/backend/auth/.env
ExecStart=$(pwd)/backend/auth/auth-server
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=pulseexpends-auth

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$(pwd)/backend/auth

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable pulseexpends-auth-server

print_status "Systemd service created and enabled"

# Create database tables
echo ""
echo "🗄️  Creating database tables..."
echo ""

cd "$AUTH_DIR"
./auth-server &
SERVER_PID=$!
sleep 5  # Wait for server to start

# Send a request to trigger auto-migration
curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/api/health

# Kill the server
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

cd - > /dev/null

print_status "Database tables created (if server started successfully)"

# Test database connection
echo ""
echo "🔍 Testing database connection..."
echo ""

if PGPASSWORD=pulseexpends_password psql -h localhost -U pulseexpends -d pulseexpends -c "SELECT 1" &> /dev/null; then
    print_status "Database connection test successful"
else
    print_warning "Database connection test failed. Please check PostgreSQL configuration."
fi

# Summary
echo ""
echo "=============================================="
echo "  Setup Complete!                            "
echo "=============================================="
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Configure Google OAuth:"
echo "   Run: ./configure-google-oauth.sh"
echo ""
echo "2. Start the authentication server:"
echo "   sudo systemctl start pulseexpends-auth-server"
echo ""
echo "3. Check server status:"
echo "   sudo systemctl status pulseexpends-auth-server"
echo ""
echo "4. View logs:"
echo "   sudo journalctl -u pulseexpends-auth-server -f"
echo ""
echo "5. Test the API:"
echo "   curl http://localhost:8082/api/health"
echo ""
echo "6. Update frontend configuration:"
echo "   Edit frontend/src/auth.js and update API_BASE_URL if needed"
echo ""
echo "📝 Important notes:"
echo "- Update GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET in backend/auth/.env"
echo "- For production, set COOKIE_SECURE=true and update CORS_ALLOWED_ORIGINS"
echo "- Consider setting up SSL/TLS for production"
echo "- Backup your .env file and database regularly"
echo ""
echo "🔐 Default database credentials:"
echo "   Database: pulseexpends"
echo "   User: pulseexpends"
echo "   Password: pulseexpends_password"
echo ""
echo "⚠️  Change the database password in production!"
echo "=============================================="