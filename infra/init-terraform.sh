#!/bin/bash

# Script to initialize Terraform for PulseExpends

echo "Initializing Terraform for PulseExpends..."

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    echo "Error: Terraform is not installed. Please install Terraform first."
    exit 1
fi

# Initialize Terraform
echo "Running terraform init..."
terraform init

if [ $? -ne 0 ]; then
    echo "Error: terraform init failed"
    exit 1
fi

# Validate configuration
echo "Running terraform validate..."
terraform validate

if [ $? -ne 0 ]; then
    echo "Error: terraform validate failed"
    exit 1
fi

# Plan the deployment
echo "Running terraform plan..."
terraform plan -out=tfplan

if [ $? -ne 0 ]; then
    echo "Error: terraform plan failed"
    exit 1
fi

echo ""
echo "✅ Terraform initialization successful!"
echo ""
echo "Next steps:"
echo "1. Review the plan: terraform show tfplan"
echo "2. Apply the deployment: terraform apply tfplan"
echo "3. After deployment, you will get the EIP address for DuckDNS configuration"
echo ""
echo "Note: You need to configure the project_id in terraform.tfvars"
echo "Current project_id: $(grep 'project_id' terraform.tfvars | head -1)"
echo ""
echo "To get your project_id from Huawei Cloud:"
echo "1. Log in to Huawei Cloud Console"
echo "2. Go to IAM > Projects"
echo "3. Find your project and copy the Project ID"
echo "4. Update the project_id in terraform.tfvars"