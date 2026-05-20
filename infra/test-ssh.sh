#!/bin/bash

# Test SSH connection to Huawei Cloud ECS instance

KEY_FILE="pulse-expends-key.pem"
IP="182.160.24.205"
INSTANCE_ID="acc9edeb-1cd4-4181-8904-881135933ac5"

echo "=== Testing SSH Connection to Huawei Cloud ECS ==="
echo "Instance: $INSTANCE_ID"
echo "IP: $IP"
echo "Key: $KEY_FILE"
echo ""

# Check if key file exists
if [ ! -f "$KEY_FILE" ]; then
    echo "❌ ERROR: Key file $KEY_FILE not found!"
    exit 1
fi

# Check key permissions
echo "🔑 Checking key permissions..."
ls -la "$KEY_FILE"
chmod 600 "$KEY_FILE" 2>/dev/null

# Test key validity
echo -e "\n🔑 Testing key validity..."
ssh-keygen -y -f "$KEY_FILE" >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Key is valid"
else
    echo "❌ Key is invalid or corrupted"
    exit 1
fi

# Try different usernames
echo -e "\n🔐 Testing SSH connection with different usernames..."
for USER in ubuntu root ec2-user admin; do
    echo -n "  Trying user '$USER'... "
    timeout 5 ssh -i "$KEY_FILE" -o ConnectTimeout=3 -o StrictHostKeyChecking=no -o BatchMode=yes "$USER@$IP" "echo 'success'" >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "✅ SUCCESS! Username is: $USER"
        exit 0
    else
        echo "❌ failed"
    fi
done

echo -e "\n🔍 Testing port 22 connectivity..."
timeout 3 nc -zv "$IP" 22 2>&1 | grep -q "succeeded"
if [ $? -eq 0 ]; then
    echo "✅ Port 22 is open"
else
    echo "❌ Port 22 is closed or unreachable"
fi

echo -e "\n📋 Next steps:"
echo "1. Go to Huawei Cloud Console → ECS → Instances"
echo "2. Find instance: $INSTANCE_ID"
echo "3. Click 'Remote Login' (VNC) to check instance status"
echo "4. Verify key pair is bound: More → Change Key Pair"
echo "5. Instance might need to be in 'Running' state"
echo ""
echo "🔄 If key binding failed, try:"
echo "   - Stop instance → Change key pair → Start instance"
echo ""
echo "🔑 Public key to import:"
echo "---"
cat "${KEY_FILE}.pub" 2>/dev/null || echo "Run: ssh-keygen -y -f $KEY_FILE > ${KEY_FILE}.pub"
echo "---"