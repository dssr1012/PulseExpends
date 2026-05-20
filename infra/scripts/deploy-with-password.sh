#!/bin/bash

# Deploy Dashboard to ECS Instance using root password
# Run this after the ECS instance is started

set -e

echo "🚀 Deploying Dashboard to ECS Instance with root password"
echo "=========================================================="

# Configuration
ECS_IP="182.160.24.205"
SSH_USER="root"
SSH_PASSWORD="dRZUO9i8NXnAhV"
DASHBOARD_DIR="/var/www/dashboard"

# Test SSH connection with password
echo "🔐 Testing SSH connection to $ECS_IP as root..."
if ! sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 "$SSH_USER@$ECS_IP" "echo 'SSH connection successful'"; then
    echo "❌ SSH connection failed!"
    echo ""
    echo "📋 Troubleshooting steps:"
    echo "   1. Check if instance is running in Huawei Cloud Console"
    echo "   2. Verify security group allows SSH (port 22)"
    echo "   3. Check if password authentication is enabled"
    echo "   4. Wait 2-3 minutes after instance start for SSH to be ready"
    exit 1
fi

echo "✅ SSH connection successful"
echo ""

# Create dashboard directory on ECS
echo "📁 Creating dashboard directory on ECS..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "
    mkdir -p $DASHBOARD_DIR
    chmod -R 755 $DASHBOARD_DIR
"

# Copy dashboard files
echo "📤 Copying dashboard files to ECS..."
sshpass -p "$SSH_PASSWORD" scp -o StrictHostKeyChecking=no dashboard.html "$SSH_USER@$ECS_IP:$DASHBOARD_DIR/"
sshpass -p "$SSH_PASSWORD" scp -o StrictHostKeyChecking=no check-status.py "$SSH_USER@$ECS_IP:$DASHBOARD_DIR/"
sshpass -p "$SSH_PASSWORD" scp -o StrictHostKeyChecking=no serve-dashboard.py "$SSH_USER@$ECS_IP:$DASHBOARD_DIR/"

# Make scripts executable on ECS
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "
    chmod +x $DASHBOARD_DIR/check-status.py
    chmod +x $DASHBOARD_DIR/serve-dashboard.py
"

# Install Python and required packages
echo "🐍 Installing Python dependencies on ECS..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "
    apt-get update
    apt-get install -y python3 python3-pip nginx
    pip3 install --upgrade pip
"

# Create systemd service for dashboard
echo "⚙️ Creating systemd service for dashboard..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "
    cat << 'EOF' > /etc/systemd/system/pulseexpends-dashboard.service
[Unit]
Description=PulseExpends Infrastructure Dashboard
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$DASHBOARD_DIR
ExecStart=/usr/bin/python3 $DASHBOARD_DIR/serve-dashboard.py
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
"

# Enable and start the service
echo "🚀 Starting dashboard service..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "
    systemctl daemon-reload
    systemctl enable pulseexpends-dashboard.service
    systemctl start pulseexpends-dashboard.service
    systemctl status pulseexpends-dashboard.service --no-pager
"

# Check if Nginx is running
echo "🌐 Checking Nginx status..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "
    if systemctl is-active --quiet nginx; then
        echo '✅ Nginx is running'
        echo '📋 Configuring Nginx to proxy dashboard...'
        
        # Create Nginx config for dashboard
        cat > /etc/nginx/sites-available/dashboard << 'NGINX_CONFIG'
server {
    listen 80;
    server_name _;
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
NGINX_CONFIG
        
        # Enable site
        ln -sf /etc/nginx/sites-available/dashboard /etc/nginx/sites-enabled/
        nginx -t && systemctl reload nginx
        echo '✅ Dashboard available at: http://$ECS_IP/'
    else
        echo '⚠️ Nginx is not running, starting it...'
        systemctl start nginx
        systemctl enable nginx
        
        # Create Nginx config for dashboard
        cat > /etc/nginx/sites-available/dashboard << 'NGINX_CONFIG'
server {
    listen 80;
    server_name _;
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
NGINX_CONFIG
        
        # Enable site
        ln -sf /etc/nginx/sites-available/dashboard /etc/nginx/sites-enabled/
        nginx -t && systemctl reload nginx
        echo '✅ Dashboard available at: http://$ECS_IP/'
    fi
"

echo ""
echo "========================================"
echo "✅ Dashboard deployment complete!"
echo ""
echo "📊 Access your dashboard at:"
echo "   - http://$ECS_IP/ (via Nginx)"
echo "   - http://$ECS_IP:8081/ (direct)"
echo ""
echo "🔧 Dashboard service commands:"
echo "   systemctl status pulseexpends-dashboard"
echo "   systemctl restart pulseexpends-dashboard"
echo "   journalctl -u pulseexpends-dashboard -f"
echo ""
echo "📈 To update the dashboard:"
echo "   1. Edit files locally"
echo "   2. Run this script again: ./deploy-with-password.sh"
echo ""
echo "🔄 To check infrastructure status:"
echo "   sshpass -p '$SSH_PASSWORD' ssh root@$ECS_IP 'cd $DASHBOARD_DIR && python3 check-status.py'"