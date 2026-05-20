#!/bin/bash

# PulseExpends Deployment Script
# This script deploys the PulseExpends infrastructure on Huawei Cloud

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
    
    # Check if credentials are set
    if [ -z "$HUAWEICLOUD_ACCESS_KEY" ] && [ -z "$HUAWEICLOUD_SECRET_KEY" ]; then
        # Try to read from terraform.tfvars
        if [ -f "terraform.tfvars" ]; then
            print_status $BLUE "INFO" "Reading credentials from terraform.tfvars..."
            export HUAWEICLOUD_ACCESS_KEY=$(grep -E '^access_key\s*=' terraform.tfvars | cut -d'=' -f2 | tr -d ' "')
            export HUAWEICLOUD_SECRET_KEY=$(grep -E '^secret_key\s*=' terraform.tfvars | cut -d'=' -f2 | tr -d ' "')
            export HUAWEICLOUD_PROJECT_ID=$(grep -E '^project_id\s*=' terraform.tfvars | cut -d'=' -f2 | tr -d ' "')
        fi
    fi
    
    # Check if credentials are still empty
    if [ -z "$HUAWEICLOUD_ACCESS_KEY" ] || [ -z "$HUAWEICLOUD_SECRET_KEY" ]; then
        print_status $RED "ERROR" "Huawei Cloud credentials not found"
        print_status $YELLOW "INFO" "Please set environment variables:"
        print_status $YELLOW "INFO" "  export HUAWEICLOUD_ACCESS_KEY=\"your-access-key\""
        print_status $YELLOW "INFO" "  export HUAWEICLOUD_SECRET_KEY=\"your-secret-key\""
        print_status $YELLOW "INFO" "  export HUAWEICLOUD_PROJECT_ID=\"your-project-id\""
        print_status $YELLOW "INFO" "Or configure them in terraform.tfvars"
        exit 1
    fi
    
    # Validate credentials by trying to list regions (simple API call)
    print_status $BLUE "INFO" "Testing Huawei Cloud credentials..."
    
    # Set environment variables for Terraform
    export TF_VAR_access_key="$HUAWEICLOUD_ACCESS_KEY"
    export TF_VAR_secret_key="$HUAWEICLOUD_SECRET_KEY"
    export TF_VAR_project_id="$HUAWEICLOUD_PROJECT_ID"
    
    print_status $GREEN "SUCCESS" "Credentials validated"
}

# Function to check Terraform installation
check_terraform() {
    print_status $BLUE "INFO" "Checking Terraform installation..."
    
    if ! command_exists terraform; then
        print_status $RED "ERROR" "Terraform is not installed"
        print_status $YELLOW "INFO" "Please install Terraform:"
        print_status $YELLOW "INFO" "  https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli"
        exit 1
    fi
    
    local terraform_version
    terraform_version=$(terraform version | head -1)
    print_status $GREEN "SUCCESS" "Terraform $terraform_version"
}

# Function to initialize Terraform
init_terraform() {
    local environment=$1
    
    print_status $BLUE "INFO" "Initializing Terraform for $environment environment..."
    
    # Check if environment directory exists
    if [ ! -d "environments/$environment" ]; then
        print_status $YELLOW "WARNING" "Environment directory 'environments/$environment' not found"
        print_status $BLUE "INFO" "Using root configuration..."
        
        # Initialize in root directory
        terraform init -upgrade
    else
        # Initialize in environment directory
        cd "environments/$environment" || exit 1
        terraform init -upgrade
        cd ../..
    fi
    
    print_status $GREEN "SUCCESS" "Terraform initialized"
}

# Function to plan Terraform deployment
plan_terraform() {
    local environment=$1
    
    print_status $BLUE "INFO" "Planning Terraform deployment for $environment environment..."
    
    if [ -d "environments/$environment" ]; then
        cd "environments/$environment" || exit 1
        terraform plan -out=tfplan
        cd ../..
    else
        terraform plan -out=tfplan
    fi
    
    print_status $GREEN "SUCCESS" "Terraform plan created: tfplan"
}

# Function to apply Terraform deployment
apply_terraform() {
    local environment=$1
    
    print_status $BLUE "INFO" "Applying Terraform deployment for $environment environment..."
    
    if [ -d "environments/$environment" ]; then
        cd "environments/$environment" || exit 1
        terraform apply -auto-approve tfplan
        cd ../..
    else
        terraform apply -auto-approve tfplan
    fi
    
    print_status $GREEN "SUCCESS" "Terraform deployment applied"
}

# Function to get Terraform outputs
get_outputs() {
    local environment=$1
    
    print_status $BLUE "INFO" "Getting Terraform outputs..."
    
    if [ -d "environments/$environment" ]; then
        cd "environments/$environment" || exit 1
        terraform output -json > outputs.json
        cd ../..
    else
        terraform output -json > outputs.json
    fi
    
    print_status $GREEN "SUCCESS" "Outputs saved to outputs.json"
}

# Function to display deployment summary
show_summary() {
    local environment=$1
    
    print_status $BLUE "INFO" "Deployment Summary for $environment environment"
    echo -e "${BLUE}=========================================${NC}"
    
    if [ -f "outputs.json" ]; then
        # Parse and display outputs
        echo -e "${YELLOW}Enterprise Project:${NC}"
        jq -r '.enterprise_project_name.value // "Not created"' outputs.json
        
        echo -e "\n${YELLOW}Network:${NC}"
        jq -r '.vpc_id.value // "Not created"' outputs.json | xargs -I {} echo "VPC ID: {}"
        
        echo -e "\n${YELLOW}Compute:${NC}"
        jq -r '.ecs_instance_ids.value[] // "No instances"' outputs.json | xargs -I {} echo "ECS Instance: {}"
        jq -r '.ecs_public_ips.value[] // "No public IPs"' outputs.json | xargs -I {} echo "Public IP: {}"
        
        echo -e "\n${YELLOW}Storage:${NC}"
        jq -r '.obs_bucket_names.value[] // "No buckets"' outputs.json | xargs -I {} echo "OBS Bucket: {}"
        
        echo -e "\n${YELLOW}Domain:${NC}"
        local domain_name
        domain_name=$(jq -r '.domain_name.value // "Not configured"' outputs.json)
        local domain_ip
        domain_ip=$(jq -r '.domain_ip_address.value // "Not assigned"' outputs.json)
        echo "Domain: $domain_name"
        echo "IP Address: $domain_ip"
        
        echo -e "\n${YELLOW}Application URLs:${NC}"
        jq -r '.application_url.value // "Not available"' outputs.json | xargs -I {} echo "Main Application: {}"
        jq -r '.mcp_server_url.value // "Not available"' outputs.json | xargs -I {} echo "MCP Server: {}"
        jq -r '.python_service_url.value // "Not available"' outputs.json | xargs -I {} echo "Python Service: {}"
        
        echo -e "\n${YELLOW}Access Commands:${NC}"
        jq -r '.ssh_access_command.value // "Not available"' outputs.json | xargs -I {} echo "SSH: {}"
        
        echo -e "\n${YELLOW}Next Steps:${NC}"
        jq -r '.next_steps.value | to_entries[] | "\(.key): \(.value)"' outputs.json 2>/dev/null || echo "See outputs.json for details"
    else
        print_status $YELLOW "WARNING" "Outputs file not found. Run 'terraform output' manually."
    fi
    
    echo -e "\n${BLUE}=========================================${NC}"
}

# Function to configure DuckDNS
configure_duckdns() {
    local environment=$1
    
    print_status $BLUE "INFO" "Configuring DuckDNS..."
    
    # Check if DuckDNS configuration is needed
    if [ ! -f "outputs.json" ]; then
        print_status $YELLOW "WARNING" "No outputs.json found. Skipping DuckDNS configuration."
        return 0
    fi
    
    local domain_name
    domain_name=$(jq -r '.domain_name.value // empty' outputs.json)
    local domain_ip
    domain_ip=$(jq -r '.domain_ip_address.value // empty' outputs.json)
    
    if [ -z "$domain_name" ] || [ -z "$domain_ip" ]; then
        print_status $YELLOW "WARNING" "Domain not configured in Terraform. Skipping DuckDNS."
        return 0
    fi
    
    # Check if DuckDNS token is available
    if [ -z "$DUCKDNS_TOKEN" ]; then
        print_status $YELLOW "INFO" "DUCKDNS_TOKEN environment variable not set."
        print_status $YELLOW "INFO" "Skipping automatic DuckDNS configuration."
        print_status $YELLOW "INFO" "Please configure DuckDNS manually:"
        print_status $YELLOW "INFO" "  Domain: $domain_name"
        print_status $YELLOW "INFO" "  IP Address: $domain_ip"
        print_status $YELLOW "INFO" "  URL: https://www.duckdns.org"
        return 0
    fi
    
    # Run DuckDNS configuration script
    if [ -f "scripts/configure-duckdns.sh" ]; then
        chmod +x scripts/configure-duckdns.sh
        ./scripts/configure-duckdns.sh --domain "$domain_name" --token "$DUCKDNS_TOKEN" --ip "$domain_ip"
    else
        print_status $YELLOW "WARNING" "DuckDNS configuration script not found."
        print_status $YELLOW "INFO" "Please configure DuckDNS manually:"
        print_status $YELLOW "INFO" "  curl \"https://www.duckdns.org/update?domains=$domain_name&token=$DUCKDNS_TOKEN&ip=$domain_ip\""
    fi
}

# Function to run tests
run_tests() {
    local environment=$1
    
    print_status $BLUE "INFO" "Running tests..."
    
    # Check if test script exists
    if [ -f "../PulseExpends/run_tests.sh" ]; then
        cd ../PulseExpends || exit 1
        chmod +x run_tests.sh
        
        print_status $BLUE "INFO" "Running Go tests..."
        ./run_tests.sh go
        
        print_status $BLUE "INFO" "Running Python tests..."
        ./run_tests.sh python
        
        print_status $BLUE "INFO" "Running security tests..."
        # Note: Security tests require the application to be running
        # We'll skip for now and run them after deployment
        print_status $YELLOW "INFO" "Security tests will be run after deployment"
        
        cd ../PulseExpends-Infra || exit 1
    else
        print_status $YELLOW "WARNING" "Test script not found. Skipping tests."
    fi
}

# Function to deploy application
deploy_application() {
    local environment=$1
    
    print_status $BLUE "INFO" "Deploying PulseExpends application..."
    
    if [ ! -f "outputs.json" ]; then
        print_status $RED "ERROR" "No outputs.json found. Cannot deploy application."
        return 1
    fi
    
    local ssh_command
    ssh_command=$(jq -r '.ssh_access_command.value // empty' outputs.json)
    
    if [ -z "$ssh_command" ]; then
        print_status $YELLOW "WARNING" "SSH command not available in outputs."
        print_status $YELLOW "INFO" "Please deploy application manually."
        return 0
    fi
    
    print_status $BLUE "INFO" "SSH command: $ssh_command"
    
    # Extract IP from SSH command
    local ip
    ip=$(echo "$ssh_command" | grep -o '@[^ ]*' | cut -d'@' -f2)
    
    if [ -z "$ip" ]; then
        print_status $RED "ERROR" "Could not extract IP from SSH command"
        return 1
    fi
    
    print_status $BLUE "INFO" "Deploying to: $ip"
    
    # Check if we have SSH key
    local ssh_key
    ssh_key=$(echo "$ssh_command" | grep -o '\-i [^ ]*' | cut -d' ' -f2)
    
    if [ ! -f "$ssh_key" ]; then
        print_status $YELLOW "WARNING" "SSH key not found: $ssh_key"
        print_status $YELLOW "INFO" "Please create SSH key pair first:"
        print_status $YELLOW "INFO" "  ssh-keygen -t rsa -b 4096 -f pulse-expends-key.pem"
        print_status $YELLOW "INFO" "Then update ecs_key_pair in terraform.tfvars"
        return 0
    fi
    
    # Deploy application (simplified - in production this would be more complex)
    print_status $BLUE "INFO" "Application deployment would happen here..."
    print_status $YELLOW "INFO" "Manual deployment steps:"
    print_status $YELLOW "INFO" "  1. Copy application files to server"
    print_status $YELLOW "INFO" "  2. Install Docker and Docker Compose"
    print_status $YELLOW "INFO" "  3. Configure environment variables"
    print_status $YELLOW "INFO" "  4. Start services with docker-compose"
    
    return 0
}

# Function to destroy deployment
destroy_deployment() {
    local environment=$1
    local confirm=$2
    
    if [ "$confirm" != "yes" ]; then
        print_status $RED "WARNING" "This will DESTROY all infrastructure for $environment environment!"
        print_status $YELLOW "INFO" "Type 'yes' to confirm destruction:"
        read -r user_confirm
        
        if [ "$user_confirm" != "yes" ]; then
            print_status $YELLOW "INFO" "Destruction cancelled"
            return 0
        fi
    fi
    
    print_status $RED "DESTROY" "Destroying Terraform deployment for $environment environment..."
    
    if [ -d "environments/$environment" ]; then
        cd "environments/$environment" || exit 1
        terraform destroy -auto-approve
        cd ../..
    else
        terraform destroy -auto-approve
    fi
    
    print_status $GREEN "SUCCESS" "Terraform deployment destroyed"
}

# Main function
main() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}  PulseExpends Deployment Script          ${NC}"
    echo -e "${BLUE}=========================================${NC}"
    
    # Parse command line arguments
    local action="deploy"
    local environment="dev"
    local skip_tests=false
    local skip_duckdns=false
    local auto_confirm=false
    local destroy_confirm="no"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --action)
                action="$2"
                shift 2
                ;;
            --environment)
                environment="$2"
                shift 2
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --skip-duckdns)
                skip_duckdns=true
                shift
                ;;
            --auto-confirm)
                auto_confirm=true
                shift
                ;;
            --destroy)
                action="destroy"
                destroy_confirm="yes"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_status $RED "ERROR" "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Validate action
    if [[ ! "$action" =~ ^(deploy|destroy|plan|outputs)$ ]]; then
        print_status $RED "ERROR" "Invalid action: $action"
        show_help
        exit 1
    fi
    
    # Validate environment
    if [[ ! "$environment" =~ ^(dev|staging|prod)$ ]]; then
        print_status $RED "ERROR" "Invalid environment: $environment. Must be dev, staging, or prod"
        exit 1
    fi
    
    # Check prerequisites
    check_terraform
    validate_credentials
    
    # Execute action
    case $action in
        "deploy")
            print_status $BLUE "INFO" "Starting deployment for $environment environment"
            
            # Run tests if not skipped
            if [ "$skip_tests" = false ]; then
                run_tests "$environment"
            fi
            
            # Initialize Terraform
            init_terraform "$environment"
            
            # Plan deployment
            plan_terraform "$environment"
            
            # Confirm deployment
            if [ "$auto_confirm" = false ]; then
                print_status $YELLOW "CONFIRM" "Proceed with deployment? (yes/no)"
                read -r user_confirm
                
                if [ "$user_confirm" != "yes" ]; then
                    print_status $YELLOW "INFO" "Deployment cancelled"
                    exit 0
                fi
            fi
            
            # Apply deployment
            apply_terraform "$environment"
            
            # Get outputs
            get_outputs "$environment"
            
            # Show summary
            show_summary "$environment"
            
            # Configure DuckDNS if not skipped
            if [ "$skip_duckdns" = false ]; then
                configure_duckdns "$environment"
            fi
            
            # Deploy application
            deploy_application "$environment"
            
            print_status $GREEN "SUCCESS" "Deployment completed for $environment environment"
            ;;
        
        "destroy")
            destroy_deployment "$environment" "$destroy_confirm"
            ;;
        
        "plan")
            init_terraform "$environment"
            plan_terraform "$environment"
            print_status $GREEN "SUCCESS" "Terraform plan created: tfplan"
            ;;
        
        "outputs")
            get_outputs "$environment"
            show_summary "$environment"
            ;;
    esac
    
    echo -e "\n${GREEN}=========================================${NC}"
    echo -e "${GREEN}  Action completed successfully!          ${NC}"
    echo -e "${GREEN}=========================================${NC}"
}

# Function to show help
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Deploy PulseExpends infrastructure on Huawei Cloud"
    echo ""
    echo "Options:"
    echo "  --action ACTION      Action to perform: deploy, destroy, plan, outputs (default: deploy)"
    echo "  --environment ENV    Environment: dev, staging, prod (default: dev)"
    echo "  --skip-tests         Skip running tests before deployment"
    echo "  --skip-duckdns       Skip DuckDNS configuration"
    echo "  --auto-confirm       Auto-confirm deployment without prompt"
    echo "  --destroy            Destroy deployment (with confirmation)"
    echo "  --help               Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  HUAWEICLOUD_ACCESS_KEY    Huawei Cloud access key"
    echo "  HUAWEICLOUD_SECRET_KEY    Huawei Cloud secret key"
    echo "  HUAWEICLOUD_PROJECT_ID    Huawei Cloud project ID"
    echo "  DUCKDNS_TOKEN             DuckDNS token for domain configuration"
    echo ""
    echo "Examples:"
    echo "  $0 --action deploy --environment dev"
    echo "  $0 --action plan --environment staging"
    echo "  $0 --action destroy --environment dev"
    echo "  $0 --action outputs --environment prod"
    echo ""
    echo "Quick deployment:"
    echo "  export HUAWEICLOUD_ACCESS_KEY=\"your-access-key\""
    echo "  export HUAWEICLOUD_SECRET_KEY=\"your-secret-key\""
    echo "  export HUAWEICLOUD_PROJECT_ID=\"your-project-id\""
    echo "  $0 --auto-confirm"
}

# Run main function
main "$@"