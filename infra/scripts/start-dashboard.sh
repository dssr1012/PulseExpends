#!/bin/bash

# PulseExpends Dashboard Launcher
# This script starts a local dashboard server and provides access instructions

set -e

echo "🚀 PulseExpends Infrastructure Dashboard"
echo "========================================"

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 is required but not installed."
    echo "   Install with: apt-get install python3"
    exit 1
fi

# Check for required files
if [ ! -f "dashboard.html" ]; then
    echo "❌ dashboard.html not found!"
    exit 1
fi

if [ ! -f "serve-dashboard.py" ]; then
    echo "❌ serve-dashboard.py not found!"
    exit 1
fi

if [ ! -f "check-status.py" ]; then
    echo "❌ check-status.py not found!"
    exit 1
fi

# Make scripts executable
chmod +x serve-dashboard.py check-status.py

# Get local IP address (for network access)
LOCAL_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")

echo ""
echo "📊 Dashboard Information:"
echo "   Local URL:    http://localhost:8081"
echo "   Network URL:  http://$LOCAL_IP:8081"
echo "   Public IP:    182.160.24.205 (ECS Instance - CURRENTLY OFFLINE)"
echo ""
echo "🔧 ECS Instance Status: SHUTOFF"
echo "   The Huawei Cloud ECS instance is currently stopped."
echo "   To start the instance and deploy the dashboard:"
echo "   1. Log into Huawei Cloud Console"
echo "   2. Go to ECS > Instances"
echo "   3. Find instance: pulseexpends-dev-pulse-afd05cc1-ecs"
echo "   4. Click 'Start'"
echo "   5. Wait 2-3 minutes for boot"
echo "   6. Run: ./deploy-dashboard-to-ecs.sh"
echo ""
echo "📈 Starting local dashboard server..."
echo "   Press Ctrl+C to stop the server"
echo ""

# Run the dashboard server
python3 serve-dashboard.py