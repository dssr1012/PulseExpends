#!/bin/bash

# Deploy Dashboard to ECS Instance
# Run this after the ECS instance is started

set -e

echo "🚀 Deploying Dashboard to ECS Instance"
echo "========================================"

# Configuration
ECS_IP="182.160.24.205"
SSH_KEY="pulse-expends-key.pem"
SSH_USER="ubuntu"
DASHBOARD_DIR="/var/www/dashboard"

# Check if SSH key exists
if [ ! -f "$SSH_KEY" ]; then
    echo "❌ SSH key not found: $SSH_KEY"
    exit 1
fi

# Check SSH key permissions
chmod 600 "$SSH_KEY"

# Test SSH connection
echo "🔐 Testing SSH connection to $ECS_IP..."
if ! ssh -i "$SSH_KEY" -o ConnectTimeout=10 -o StrictHostKeyChecking=no "$SSH_USER@$ECS_IP" "echo 'SSH connection successful'"; then
    echo "❌ SSH connection failed!"
    echo ""
    echo "📋 Troubleshooting steps:"
    echo "   1. Check if instance is running in Huawei Cloud Console"
    echo "   2. Verify security group allows SSH (port 22)"
    echo "   3. Check if key pair 'pulse-expends-key' is attached to instance"
    echo "   4. Wait 2-3 minutes after instance start for SSH to be ready"
    echo ""
    echo "💡 To start the instance:"
    echo "   - Log into Huawei Cloud Console"
    echo "   - Go to ECS > Instances"
    echo "   - Find: pulseexpends-dev-pulse-afd05cc1-ecs"
    echo "   - Click 'Start'"
    exit 1
fi

echo "✅ SSH connection successful"
echo ""

# Create dashboard directory on ECS
echo "📁 Creating dashboard directory on ECS..."
ssh -i "$SSH_KEY" "$SSH_USER@$ECS_IP" "
    sudo mkdir -p $DASHBOARD_DIR
    sudo chown -R $SSH_USER:$SSH_USER $DASHBOARD_DIR
    sudo chmod -R 755 $DASHBOARD_DIR
"

# Copy dashboard files
echo "📤 Copying dashboard files to ECS..."
scp -i "$SSH_KEY" dashboard.html "$SSH_USER@$ECS_IP:$DASHBOARD_DIR/"
scp -i "$SSH_KEY" check-status.py "$SSH_USER@$ECS_IP:$DASHBOARD_DIR/"
scp -i "$SSH_KEY" serve-dashboard.py "$SSH_USER@$ECS_IP:$DASHBOARD_DIR/"

# Make scripts executable on ECS
ssh -i "$SSH_KEY" "$SSH_USER@$ECS_IP" "
    chmod +x $DASHBOARD_DIR/check-status.py
    chmod +x $DASHBOARD_DIR/serve-dashboard.py
"

# Install Python and required packages
echo "🐍 Installing Python dependencies on ECS..."
ssh -i "$SSH_KEY" "$SSH_USER@$ECS_IP" "
    sudo apt-get update
    sudo apt-get install -y python3 python3-pip
    pip3 install --upgrade pip
"

# Create systemd service for dashboard
echo "⚙️ Creating systemd service for dashboard..."
ssh -i "$SSH_KEY" "$SSH_USER@$ECS_IP" "
    cat << EOF | sudo tee /etc/systemd/system/pulseexpends-dashboard.service
[Unit]
Description=PulseExpends Infrastructure Dashboard
After=network.target

[Service]
Type=simple
User=$SSH_USER
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
ssh -i "$SSH_KEY" "$SSH_USER@$ECS_IP" "
    sudo systemctl daemon-reload
    sudo systemctl enable pulseexpends-dashboard.service
    sudo systemctl start pulseexpends-dashboard.service
    sudo systemctl status pulseexpends-dashboard.service --no-pager
"

# Check if Nginx is running (from user-data script)
echo "🌐 Checking Nginx status..."
ssh -i "$SSH_KEY" "$SSH_USER@$ECS_IP" "
    if systemctl is-active --quiet nginx; then
        echo '✅ Nginx is running'
        echo '📋 Configuring Nginx to proxy dashboard...'
        
        # Create Nginx config for dashboard
        sudo tee /etc/nginx/sites-available/dashboard << 'NGINX_CONFIG'
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
        sudo ln -sf /etc/nginx/sites-available/dashboard /etc/nginx/sites-enabled/
        sudo nginx -t && sudo systemctl reload nginx
        echo '✅ Dashboard available at: http://$ECS_IP/'
    else
        echo '⚠️ Nginx is not running'
        echo '📋 Dashboard available at: http://$ECS_IP:8081/'
    fi
"

echo ""
echo "========================================"
echo "✅ Dashboard deployment complete!"
echo ""
echo "📊 Access your dashboard at:"
echo "   - http://$ECS_IP/ (if Nginx is running)"
echo "   - http://$ECS_IP:8081/ (direct)"
echo ""
echo "🔧 Dashboard service commands:"
echo "   sudo systemctl status pulseexpends-dashboard"
echo "   sudo systemctl restart pulseexpends-dashboard"
echo "   sudo journalctl -u pulseexpends-dashboard -f"
echo ""
echo "📈 To update the dashboard:"
echo "   1. Edit files locally"
echo "   2. Run this script again: ./deploy-dashboard-to-ecs.sh"
echo ""
echo "🔄 To check infrastructure status:"
echo "   ssh -i $SSH_KEY $SSH_USER@$ECS_IP 'cd $DASHBOARD_DIR && python3 check-status.py'"