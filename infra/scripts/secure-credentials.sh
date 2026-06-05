#!/bin/bash

# PulseExpends Secure Credential Management Script
# This script helps manage Huawei Cloud credentials securely

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local status=$2
    local message=$3
    echo -e "${color}[${status}]${NC} ${message}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to validate Huawei Cloud credentials
validate_credentials() {
    print_status $BLUE "INFO" "Validating Huawei Cloud credentials..."
    
    if [ -z "$HUAWEICLOUD_ACCESS_KEY" ]; then
        print_status $RED "ERROR" "HUAWEICLOUD_ACCESS_KEY environment variable is not set"
        exit 1
    fi
    
    if [ -z "$HUAWEICLOUD_SECRET_KEY" ]; then
        print_status $RED "ERROR" "HUAWEICLOUD_SECRET_KEY environment variable is not set"
        exit 1
    fi
    
    if [ -z "$HUAWEICLOUD_PROJECT_ID" ]; then
        print_status $RED "ERROR" "HUAWEICLOUD_PROJECT_ID environment variable is not set"
        exit 1
    fi
    
    print_status $GREEN "SUCCESS" "Huawei Cloud credentials are set in environment variables"
}

# Function to generate secure terraform.tfvars
generate_secure_tfvars() {
    print_status $BLUE "INFO" "Generating secure terraform.tfvars..."
    
    # Check if terraform.tfvars already exists
    if [ -f "terraform.tfvars" ]; then
        print_status $YELLOW "WARNING" "terraform.tfvars already exists"
        read -p "Do you want to backup and overwrite? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status $YELLOW "INFO" "Keeping existing terraform.tfvars"
            return 0
        fi
        
        # Create backup
        BACKUP_FILE="terraform.tfvars.backup.$(date +%Y%m%d%H%M%S)"
        cp terraform.tfvars "$BACKUP_FILE"
        print_status $GREEN "SUCCESS" "Backup created: $BACKUP_FILE"
    fi
    
    # Generate secure terraform.tfvars
    cat > terraform.tfvars << EOF
# ============================================================================
# Huawei Cloud Credentials
# ============================================================================
# WARNING: This file contains sensitive information
# DO NOT commit to version control
# Use environment variables or a secrets manager instead

# Required: Huawei Cloud credentials
access_key = "$HUAWEICLOUD_ACCESS_KEY"
secret_key = "$HUAWEICLOUD_SECRET_KEY"
project_id = "$HUAWEICLOUD_PROJECT_ID"

# ============================================================================
# Region & Environment Configuration
# ============================================================================

# Huawei Cloud region (la-south-2 = Santiago, Chile)
region = "la-south-2"

# Environment name (dev, staging, prod)
environment = "dev"

# Project name for resource naming
project_name = "pulseexpends"

# ============================================================================
# Enterprise Project Configuration
# ============================================================================

# Enable Enterprise Project (requires Enterprise Project permissions)
enable_enterprise_project = false

# Enterprise Project name (only used if enable_enterprise_project = true)
enterprise_project_name = "pulse-expends"

# Enterprise Project type (prod, poc, dev)
enterprise_project_type = "prod"

# ============================================================================
# Domain Configuration
# ============================================================================

# Enable custom domain
enable_domain = true

# Domain name (e.g., pulseexpends.duckdns.org)
domain_name = "pulseexpends.duckdns.org"

# Set to true if using Huawei Cloud DNS, false for external DNS like DuckDNS
create_dns_record = false

# DNS zone ID (required if create_dns_record is true)
dns_zone_id = ""

# Maximum bandwidth size in Mbps (bandwidth cap when using traffic billing)
domain_bandwidth_size = 100

# ============================================================================
# SSL Configuration
# ============================================================================

# Enable SSL/TLS (requires valid certificate)
enable_ssl = false

# SSL certificate content (PEM format)
# ssl_certificate = ""

# SSL private key content (PEM format)
# ssl_private_key = ""

# ============================================================================
# Network Configuration
# ============================================================================

# VPC CIDR block
vpc_cidr = "10.0.0.0/16"

# Public subnets (for ECS instances with public IPs)
public_subnets = ["10.0.1.0/24"]

# Private subnets (for RDS and internal services)
private_subnets = ["10.0.101.0/24"]

# Enable NAT Gateway for private subnets
enable_nat_gateway = false

# Use single NAT Gateway for all private subnets
single_nat_gateway = true

# Enable VPC Flow Logs (for network traffic monitoring)
enable_flow_logs = true

# OBS bucket name for VPC Flow Logs
flow_log_bucket = "pulseexpends-flow-logs"

# ============================================================================
# Compute Configuration (Optimized for 10 concurrent users)
# ============================================================================

# ECS instance type (optimized for 10 users)
ecs_instance_type = "s6.large.2"

# Number of ECS instances
ecs_instance_count = 1

# ECS instance disk size (GB)
ecs_disk_size = 40

# ECS instance disk type
ecs_disk_type = "SSD"

# ECS instance image ID (Ubuntu 22.04 LTS)
ecs_image_id = ""

# ECS instance key pair name
ecs_key_pair = "pulse-expends-key"

# ============================================================================
# Database Configuration (RDS PostgreSQL)
# ============================================================================

# Enable RDS PostgreSQL database
enable_rds = true

# RDS instance type (optimized for 10 users)
rds_instance_type = "rds.pg.n1.large.2"

# RDS engine version
rds_engine_version = "15"

# RDS storage size (GB)
rds_storage = 50

# RDS backup window (UTC)
rds_backup_window = "02:00-03:00"

# RDS backup retention period (days)
rds_backup_retention = 7

# RDS database name
rds_database_name = "pulseexpends"

# RDS username
rds_username = "pulseexpends_admin"

# RDS password (auto-generated if empty)
rds_password = "$(openssl rand -base64 32)"

# RDS HA replication mode (async or semisync)
rds_ha_replication_mode = "async"

# ============================================================================
# Storage Configuration (OBS)
# ============================================================================

# Enable OBS bucket for document storage
enable_obs = true

# OBS bucket name
obs_bucket_name = "pulseexpends-documents"

# OBS bucket storage class (STANDARD, WARM, COLD)
obs_storage_class = "STANDARD"

# OBS bucket versioning
obs_versioning = true

# ============================================================================
# Load Balancer Configuration
# ============================================================================

# Enable Elastic Load Balancer
enable_elb = false

# ELB bandwidth size (Mbps)
elb_bandwidth_size = 5

# ELB listener protocol (HTTP or HTTPS)
elb_listener_protocol = "HTTP"

# ELB listener port
elb_listener_port = 80

# ============================================================================
# Monitoring Configuration
# ============================================================================

# Enable Cloud Eye monitoring
enable_monitoring = true

# Cloud Eye alarm notification email
monitoring_email = ""

# Cloud Eye alarm notification phone
monitoring_phone = ""

# ============================================================================
# Security Configuration
# ============================================================================

# Allowed SSH IP ranges (restrict to specific IPs)
allowed_ssh_ips = ["0.0.0.0/0"]  # WARNING: Change this to your IP!

# Allowed HTTP IP ranges
allowed_http_ips = ["0.0.0.0/0"]

# Allowed HTTPS IP ranges
allowed_https_ips = ["0.0.0.0/0"]

# ============================================================================
# Tags
# ============================================================================

# Additional tags for all resources
additional_tags = {
  Owner       = "DevOps"
  Department  = "Engineering"
  CostCenter  = "PulseExpends"
  Environment = "Development"
}
EOF
    
    print_status $GREEN "SUCCESS" "Generated secure terraform.tfvars"
    
    # Set restrictive permissions
    chmod 600 terraform.tfvars
    print_status $GREEN "SUCCESS" "Set restrictive permissions (600) on terraform.tfvars"
}

# Function to setup environment variables
setup_env_vars() {
    print_status $BLUE "INFO" "Setting up environment variables..."
    
    # Create .env file template
    cat > .env.example << EOF
# Huawei Cloud Credentials
export HUAWEICLOUD_ACCESS_KEY="your_access_key"
export HUAWEICLOUD_SECRET_KEY="your_secret_key"
export HUAWEICLOUD_PROJECT_ID="your_project_id"

# Terraform Variables
export TF_VAR_access_key="\$HUAWEICLOUD_ACCESS_KEY"
export TF_VAR_secret_key="\$HUAWEICLOUD_SECRET_KEY"
export TF_VAR_project_id="\$HUAWEICLOUD_PROJECT_ID"

# Optional: RDS Password
export TF_VAR_rds_password="\$(openssl rand -base64 32)"

# Optional: Monitoring Email
export TF_VAR_monitoring_email="your-email@example.com"
EOF
    
    print_status $GREEN "SUCCESS" "Created .env.example template"
    print_status $YELLOW "INFO" "Copy .env.example to .env and update with your credentials"
    print_status $YELLOW "INFO" "Then run: source .env"
}

# Function to check for hardcoded credentials
check_hardcoded_credentials() {
    print_status $BLUE "INFO" "Checking for hardcoded credentials..."
    
    local found_creds=false
    
    # Check terraform.tfvars for hardcoded credentials
    if [ -f "terraform.tfvars" ]; then
        if grep -q "access_key = \"" terraform.tfvars && ! grep -q "access_key = \"YOUR_ACCESS_KEY_HERE\"" terraform.tfvars; then
            print_status $RED "WARNING" "Found hardcoded access_key in terraform.tfvars"
            found_creds=true
        fi
        
        if grep -q "secret_key = \"" terraform.tfvars && ! grep -q "secret_key = \"YOUR_SECRET_KEY_HERE\"" terraform.tfvars; then
            print_status $RED "WARNING" "Found hardcoded secret_key in terraform.tfvars"
            found_creds=true
        fi
    fi
    
    # Check for other sensitive files
    for file in *.tfvars *.tf; do
        if [ -f "$file" ]; then
            if grep -q -E "(access_key|secret_key|password|token)" "$file" 2>/dev/null; then
                print_status $YELLOW "WARNING" "Potential sensitive data in $file"
                grep -n -E "(access_key|secret_key|password|token)" "$file" | head -5
            fi
        fi
    done
    
    if [ "$found_creds" = false ]; then
        print_status $GREEN "SUCCESS" "No hardcoded credentials found"
    else
        print_status $RED "ERROR" "Hardcoded credentials found. Please use environment variables instead."
        exit 1
    fi
}

# Function to setup Git hooks for security
setup_git_hooks() {
    print_status $BLUE "INFO" "Setting up Git hooks for security..."
    
    # Create pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Pre-commit hook to prevent committing sensitive data

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Checking for sensitive data..."

# Patterns to check for
PATTERNS=(
    "access_key\s*="
    "secret_key\s*="
    "password\s*="
    "token\s*="
    "private_key\s*="
    "certificate\s*="
    "BEGIN RSA PRIVATE KEY"
    "BEGIN PRIVATE KEY"
    "BEGIN CERTIFICATE"
)

FILES_TO_CHECK=$(git diff --cached --name-only)

for file in $FILES_TO_CHECK; do
    for pattern in "${PATTERNS[@]}"; do
        if git diff --cached "$file" | grep -q -E "$pattern"; then
            echo -e "${RED}ERROR: Potential sensitive data found in $file${NC}"
            echo -e "${RED}Pattern: $pattern${NC}"
            exit 1
        fi
    done
done

echo -e "${GREEN}No sensitive data found.${NC}"
exit 0
EOF
    
    chmod +x .git/hooks/pre-commit
    print_status $GREEN "SUCCESS" "Git pre-commit hook installed"
}

# Function to create IAM policy for least privilege
create_iam_policy() {
    print_status $BLUE "INFO" "Creating IAM policy for least privilege access..."
    
    cat > iam-policy.json << 'EOF'
{
  "Version": "1.1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "vpc:*VPC*",
        "vpc:*Subnet*",
        "vpc:*SecurityGroup*",
        "vpc:*EIP*",
        "ecs:*",
        "rds:*",
        "obs:*",
        "ces:*",
        "kms:*",
        "iam:Get*",
        "iam:List*",
        "iam:CreateAgency",
        "iam:DeleteAgency",
        "iam:UpdateAgency"
      ],
      "Resource": ["*"]
    },
    {
      "Effect": "Deny",
      "Action": [
        "iam:CreateUser",
        "iam:DeleteUser",
        "iam:CreateGroup",
        "iam:DeleteGroup",
        "kms:ScheduleKeyDeletion",
        "rds:DeleteDBInstance"
      ],
      "Resource": ["*"]
    }
  ]
}
EOF
    
    print_status $GREEN "SUCCESS" "Created IAM policy template: iam-policy.json"
    print_status $YELLOW "INFO" "Apply this policy to your IAM user/group for least privilege access"
}

# Main function
main() {
    echo "========================================="
    echo "PulseExpends Secure Credential Management"
    echo "========================================="
    echo ""
    
    case "${1:-}" in
        "validate")
            validate_credentials
            ;;
        "generate")
            validate_credentials
            generate_secure_tfvars
            ;;
        "setup-env")
            setup_env_vars
            ;;
        "check")
            check_hardcoded_credentials
            ;;
        "git-hooks")
            setup_git_hooks
            ;;
        "iam-policy")
            create_iam_policy
            ;;
        "all")
            validate_credentials
            check_hardcoded_credentials
            generate_secure_tfvars
            setup_env_vars
            setup_git_hooks
            create_iam_policy
            ;;
        *)
            echo "Usage: $0 {validate|generate|setup-env|check|git-hooks|iam-policy|all}"
            echo ""
            echo "Commands:"
            echo "  validate     - Validate Huawei Cloud credentials"
            echo "  generate     - Generate secure terraform.tfvars from env vars"
            echo "  setup-env    - Setup environment variables template"
            echo "  check        - Check for hardcoded credentials"
            echo "  git-hooks    - Install Git hooks for security"
            echo "  iam-policy   - Create IAM policy for least privilege"
            echo "  all          - Run all security setup steps"
            echo ""
            echo "Example:"
            echo "  export HUAWEICLOUD_ACCESS_KEY=\"your_key\""
            echo "  export HUAWEICLOUD_SECRET_KEY=\"your_secret\""
            echo "  export HUAWEICLOUD_PROJECT_ID=\"your_project\""
            echo "  $0 generate"
            exit 1
            ;;
    esac
    
    echo ""
    print_status $GREEN "DONE" "Secure credential management completed"
}

# Run main function
main "$@"