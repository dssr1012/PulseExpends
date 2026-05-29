#!/bin/bash
# Test SSH connectivity to ECS instance

IP="182.160.24.205"
KEY_FILE="pulse-expends-key.pem"

echo "🔍 Testing SSH connectivity to ECS instance at $IP..."
echo "Using SSH key: $KEY_FILE"

# Check if key file exists
if [ ! -f "$KEY_FILE" ]; then
    echo "❌ SSH key file not found: $KEY_FILE"
    echo "Generating new key pair..."
    ssh-keygen -t rsa -b 4096 -f "$KEY_FILE" -N "" -q
    echo "✅ SSH key generated"
fi

# Test SSH connection
echo "🔌 Testing SSH connection..."
ssh -i "$KEY_FILE" -o ConnectTimeout=10 -o StrictHostKeyChecking=no -o BatchMode=yes "root@$IP" "echo 'SSH connection successful'" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✅ SSH connection successful!"
    echo ""
    echo "🚀 ECS instance is running and accessible"
    echo "   IP Address: $IP"
    echo "   SSH Key: $KEY_FILE"
    echo ""
    echo "📋 To deploy PulseExpends:"
    echo "   1. Run: chmod +x deploy-app.sh"
    echo "   2. Run: ./deploy-app.sh"
else
    echo "❌ SSH connection failed"
    echo ""
    echo "📋 Possible issues:"
    echo "   1. ECS instance is stopped"
    echo "   2. SSH key not bound to instance"
    echo "   3. Security group blocking port 22"
    echo "   4. Wrong IP address"
    echo ""
    echo "🔧 Troubleshooting steps:"
    echo "   1. Check Huawei Cloud Console → ECS → Instances"
    echo "   2. Ensure instance is in 'ACTIVE' state"
    echo "   3. Check security group rules allow SSH (port 22)"
    echo "   4. Verify the IP address is correct"
    echo "   5. Import SSH public key to Huawei Cloud:"
    echo ""
    echo "📋 Public key to import:"
    echo "------------------------------------------------------------"
    ssh-keygen -y -f "$KEY_FILE"
    echo "------------------------------------------------------------"
    echo ""
    echo "💡 After importing key and starting instance, run this script again"
fi