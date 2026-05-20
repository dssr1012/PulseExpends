#!/bin/bash

# Script to get Huawei Cloud project ID

echo "Getting Huawei Cloud project ID for region: la-south-2"

# Check if Huawei Cloud CLI is installed
if ! command -v huaweicloud &> /dev/null; then
    echo "Huawei Cloud CLI not found. Installing..."
    
    # Install Huawei Cloud CLI
    curl -sSL https://hwcloudcli.obs.cn-north-1.myhuaweicloud.com/cli/latest/huaweicloud-cli-linux-amd64.tar.gz -o huaweicloud-cli.tar.gz
    tar -xzf huaweicloud-cli.tar.gz
    sudo mv huaweicloud /usr/local/bin/
    rm huaweicloud-cli.tar.gz
    
    echo "Huawei Cloud CLI installed."
fi

# Configure Huawei Cloud CLI
echo "Configuring Huawei Cloud CLI..."
read -p "Enter your Access Key ID: " ACCESS_KEY
read -p "Enter your Secret Access Key: " -s SECRET_KEY
echo ""
read -p "Enter region (default: la-south-2): " REGION
REGION=${REGION:-la-south-2}

# Configure credentials
huaweicloud configure set --access-key-id "$ACCESS_KEY" --secret-access-key "$SECRET_KEY" --region "$REGION"

if [ $? -ne 0 ]; then
    echo "Error: Failed to configure Huawei Cloud CLI"
    exit 1
fi

# Get project ID
echo "Fetching project information..."
PROJECT_INFO=$(huaweicloud iam project list --name "My Project" 2>/dev/null || huaweicloud iam project list)

if [ $? -ne 0 ]; then
    echo "Error: Failed to get project list"
    echo "Please manually get your project ID from Huawei Cloud Console:"
    echo "1. Log in to https://console.huaweicloud.com/"
    echo "2. Go to IAM > Projects"
    echo "3. Find your project in region $REGION"
    echo "4. Copy the Project ID"
    exit 1
fi

# Try to parse project ID
PROJECT_ID=$(echo "$PROJECT_INFO" | grep -oP '(?<=project_id": ")[^"]+' | head -1)

if [ -z "$PROJECT_ID" ]; then
    echo "Could not automatically find project ID."
    echo ""
    echo "Please manually get your project ID:"
    echo "1. Log in to Huawei Cloud Console"
    echo "2. Go to IAM > Projects"
    echo "3. Find your project in region $REGION"
    echo "4. Copy the Project ID"
    echo ""
    echo "Then update terraform.tfvars:"
    echo "project_id = \"YOUR_PROJECT_ID_HERE\""
else
    echo ""
    echo "✅ Found Project ID: $PROJECT_ID"
    echo ""
    echo "Updating terraform.tfvars..."
    
    # Update terraform.tfvars
    if [ -f "terraform.tfvars" ]; then
        sed -i "s/project_id = \".*\"/project_id = \"$PROJECT_ID\"/" terraform.tfvars
        echo "Updated terraform.tfvars with project_id: $PROJECT_ID"
    else
        echo "Error: terraform.tfvars not found"
        echo "Please manually set project_id = \"$PROJECT_ID\" in terraform.tfvars"
    fi
    
    echo ""
    echo "Next steps:"
    echo "1. Verify the project_id in terraform.tfvars"
    echo "2. Run: ./init-terraform.sh"
    echo "3. Run: terraform apply"
fi

# Clean up credentials from memory
unset ACCESS_KEY
unset SECRET_KEY