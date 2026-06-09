#!/bin/bash
# PulseExpends RDS Configuration Script
# Updates .env file with RDS PostgreSQL 17 connection details

ENV_FILE="/root/PulseExpends/services/backend/mcp-server-integrated/.env"
BACKUP_FILE="${ENV_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

echo "=== PulseExpends RDS Configuration Script ==="
echo "Target file: $ENV_FILE"

# Backup existing .env file if it exists
if [ -f "$ENV_FILE" ]; then
    cp "$ENV_FILE" "$BACKUP_FILE"
    echo "Backup created: $BACKUP_FILE"
else
    echo "Creating new .env file"
fi

# Create/update .env file with RDS configuration
cat > "$ENV_FILE" << 'EOF'
# Database Configuration
DATABASE_URL=postgres://root:YOUR_PASSWORD@10.0.101.232:5432/pulseexpends_auth?sslmode=disable

# Server Configuration
PORT=3001
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=24h

# Logging
LOG_LEVEL=debug

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://159.138.118.60

# Application
APP_ENV=production
EOF

echo ""
echo "=== Configuration Applied ==="
echo "Host: 10.0.101.232"
echo "Port: 5432"
echo "Database: pulseexpends_auth"
echo "User: root"
echo "Password: YOUR_PASSWORD (REPLACE THIS MANUALLY)"
echo ""
echo "Please replace YOUR_PASSWORD with the actual RDS password in $ENV_FILE"
echo "Then run: source $ENV_FILE"
echo ""

# Set permissions
chmod 600 "$ENV_FILE"
echo "File permissions updated (read/write owner only)"

# Display the file
echo ""
echo "=== Current .env Contents ==="
cat "$ENV_FILE"