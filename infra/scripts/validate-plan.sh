#!/bin/bash

# PulseExpends Terraform Validation Script
# Validates Terraform configuration before deployment

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

# Function to install required tools
install_tools() {
    print_status $BLUE "INFO" "Checking required tools..."
    
    # Check Terraform
    if ! command_exists terraform; then
        print_status $RED "ERROR" "Terraform is not installed"
        print_status $YELLOW "INFO" "Install Terraform: https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli"
        exit 1
    fi
    
    # Check tfsec
    if ! command_exists tfsec; then
        print_status $YELLOW "WARNING" "tfsec is not installed (security scanning)"
        print_status $YELLOW "INFO" "Install with: brew install tfsec  # macOS"
        print_status $YELLOW "INFO" "Or: curl -s https://raw.githubusercontent.com/aquasecurity/tfsec/master/scripts/install_linux.sh | bash"
    fi
    
    # Check checkov
    if ! command_exists checkov; then
        print_status $YELLOW "WARNING" "checkov is not installed (security scanning)"
        print_status $YELLOW "INFO" "Install with: pip install checkov"
    fi
    
    # Check infracost
    if ! command_exists infracost; then
        print_status $YELLOW "WARNING" "infracost is not installed (cost estimation)"
        print_status $YELLOW "INFO" "Install with: brew install infracost  # macOS"
        print_status $YELLOW "INFO" "Or: curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh"
    fi
    
    print_status $GREEN "SUCCESS" "Required tools checked"
}

# Function to validate Terraform syntax
validate_syntax() {
    print_status $BLUE "INFO" "Validating Terraform syntax..."
    
    # Format check
    print_status $BLUE "INFO" "Checking Terraform format..."
    if terraform fmt -check -recursive; then
        print_status $GREEN "SUCCESS" "Terraform format is correct"
    else
        print_status $YELLOW "WARNING" "Terraform format issues found"
        print_status $YELLOW "INFO" "Run: terraform fmt -recursive"
    fi
    
    # Validate configuration
    print_status $BLUE "INFO" "Validating Terraform configuration..."
    if terraform validate; then
        print_status $GREEN "SUCCESS" "Terraform configuration is valid"
    else
        print_status $RED "ERROR" "Terraform configuration validation failed"
        exit 1
    fi
}

# Function to run security scans
run_security_scans() {
    print_status $BLUE "INFO" "Running security scans..."
    
    # tfsec scan
    if command_exists tfsec; then
        print_status $BLUE "INFO" "Running tfsec security scan..."
        tfsec . --verbose
        if [ $? -eq 0 ]; then
            print_status $GREEN "SUCCESS" "tfsec scan passed"
        else
            print_status $YELLOW "WARNING" "tfsec found security issues"
        fi
    fi
    
    # checkov scan
    if command_exists checkov; then
        print_status $BLUE "INFO" "Running checkov security scan..."
        checkov --directory . --quiet
        if [ $? -eq 0 ]; then
            print_status $GREEN "SUCCESS" "checkov scan passed"
        else
            print_status $YELLOW "WARNING" "checkov found security issues"
        fi
    fi
}

# Function to estimate costs
estimate_costs() {
    print_status $BLUE "INFO" "Estimating infrastructure costs..."
    
    if command_exists infracost; then
        # Check if infracost is configured
        if [ -f "~/.config/infracost/credentials.yml" ] || [ -n "$INFRACOST_API_KEY" ]; then
            print_status $BLUE "INFO" "Running cost estimation..."
            infracost breakdown --path . --format table --show-skipped
        else
            print_status $YELLOW "INFO" "Infracost not configured. Set INFRACOST_API_KEY environment variable"
            print_status $YELLOW "INFO" "Get API key from: https://dashboard.infracost.io"
        fi
    fi
}

# Function to generate execution plan
generate_plan() {
    print_status $BLUE "INFO" "Generating Terraform execution plan..."
    
    # Check if credentials are set
    if [ -z "$TF_VAR_access_key" ] && [ -z "$HUAWEICLOUD_ACCESS_KEY" ]; then
        print_status $RED "ERROR" "Huawei Cloud credentials not set"
        print_status $YELLOW "INFO" "Set environment variables:"
        print_status $YELLOW "INFO" "  export TF_VAR_access_key=\"your_access_key\""
        print_status $YELLOW "INFO" "  export TF_VAR_secret_key=\"your_secret_key\""
        print_status $YELLOW "INFO" "  export TF_VAR_project_id=\"your_project_id\""
        exit 1
    fi
    
    # Initialize Terraform if not already initialized
    if [ ! -d ".terraform" ]; then
        print_status $BLUE "INFO" "Initializing Terraform..."
        terraform init
    fi
    
    # Generate plan
    PLAN_FILE="tfplan-$(date +%Y%m%d%H%M%S).out"
    print_status $BLUE "INFO" "Generating plan file: $PLAN_FILE"
    
    terraform plan -out="$PLAN_FILE" -detailed-exitcode
    EXIT_CODE=$?
    
    case $EXIT_CODE in
        0)
            print_status $GREEN "SUCCESS" "No changes needed. Infrastructure is up-to-date."
            ;;
        1)
            print_status $RED "ERROR" "Terraform plan failed"
            exit 1
            ;;
        2)
            print_status $YELLOW "WARNING" "Changes detected. Review plan below:"
            echo ""
            terraform show "$PLAN_FILE" | grep -E "(Plan:|~|\+|\-)" | head -50
            echo ""
            print_status $BLUE "INFO" "Full plan saved to: $PLAN_FILE"
            print_status $BLUE "INFO" "Review with: terraform show $PLAN_FILE"
            print_status $BLUE "INFO" "Apply with: terraform apply $PLAN_FILE"
            ;;
        *)
            print_status $RED "ERROR" "Unexpected exit code: $EXIT_CODE"
            exit 1
            ;;
    esac
}

# Function to check for destructive changes
check_destructive_changes() {
    print_status $BLUE "INFO" "Checking for destructive changes..."
    
    if [ -f "$PLAN_FILE" ]; then
        DESTRUCTIVE_COUNT=$(terraform show -json "$PLAN_FILE" | jq -r '.resource_changes[] | select(.change.actions[] | contains("delete")) | .address' | wc -l)
        
        if [ "$DESTRUCTIVE_COUNT" -gt 0 ]; then
            print_status $RED "WARNING" "Found $DESTRUCTIVE_COUNT destructive change(s):"
            terraform show -json "$PLAN_FILE" | jq -r '.resource_changes[] | select(.change.actions[] | contains("delete")) | "  - \(.address)"'
            echo ""
            print_status $YELLOW "INFO" "Review these changes carefully before applying!"
        else
            print_status $GREEN "SUCCESS" "No destructive changes detected"
        fi
    fi
}

# Function to validate resource limits
validate_resource_limits() {
    print_status $BLUE "INFO" "Validating resource limits for 10 concurrent users..."
    
    # Check ECS instance type
    ECS_TYPE=$(grep -E 'ecs_instance_type\s*=' terraform.tfvars 2>/dev/null || echo 'ecs_instance_type = "s6.large.2"')
    if [[ ! "$ECS_TYPE" =~ s6\.large\.2 ]]; then
        print_status $YELLOW "WARNING" "ECS instance type may not be optimized for 10 users"
        print_status $YELLOW "INFO" "Recommended: s6.large.2 (2 vCPUs, 4GB RAM)"
    fi
    
    # Check RDS instance type
    RDS_TYPE=$(grep -E 'rds_instance_type\s*=' terraform.tfvars 2>/dev/null || echo 'rds_instance_type = "rds.pg.n1.large.2"')
    if [[ ! "$RDS_TYPE" =~ rds\.pg\.n1\.large\.2 ]]; then
        print_status $YELLOW "WARNING" "RDS instance type may not be optimized for 10 users"
        print_status $YELLOW "INFO" "Recommended: rds.pg.n1.large.2 (2 vCPUs, 4GB RAM)"
    fi
    
    # Check instance count
    INSTANCE_COUNT=$(grep -E 'ecs_instance_count\s*=' terraform.tfvars 2>/dev/null || echo 'ecs_instance_count = 1')
    if [[ ! "$INSTANCE_COUNT" =~ 1 ]]; then
        print_status $YELLOW "WARNING" "Multiple ECS instances may be unnecessary for 10 users"
        print_status $YELLOW "INFO" "Recommended: 1 instance for cost optimization"
    fi
    
    print_status $GREEN "SUCCESS" "Resource limits validated"
}

# Function to generate validation report
generate_report() {
    print_status $BLUE "INFO" "Generating validation report..."
    
    REPORT_FILE="validation-report-$(date +%Y%m%d%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# Terraform Validation Report
## PulseExpends Infrastructure - $(date)

### Summary
- **Timestamp**: $(date)
- **Terraform Version**: $(terraform version | head -1)
- **Validation Status**: $(if [ $OVERALL_STATUS -eq 0 ]; then echo "PASSED"; else echo "FAILED"; fi)

### Security Scans
$(if command_exists tfsec; then
  echo "- **tfsec**: $(tfsec . --format json | jq -r '.results | length') issues found"
else
  echo "- **tfsec**: Not installed"
fi)

$(if command_exists checkov; then
  echo "- **checkov**: $(checkov --directory . --quiet --output json | jq -r '.summary.failed') checks failed"
else
  echo "- **checkov**: Not installed"
fi)

### Cost Estimation
$(if command_exists infracost && ([ -f "~/.config/infracost/credentials.yml" ] || [ -n "$INFRACOST_API_KEY" ]); then
  echo "- **Monthly Cost**: Estimated"
else
  echo "- **Monthly Cost**: Not estimated (infracost not configured)"
fi)

### Resource Validation
- **ECS Instance**: $(echo "$ECS_TYPE" | cut -d'=' -f2 | tr -d ' "\t')
- **RDS Instance**: $(echo "$RDS_TYPE" | cut -d'=' -f2 | tr -d ' "\t')
- **Instance Count**: $(echo "$INSTANCE_COUNT" | cut -d'=' -f2 | tr -d ' "\t')

### Recommendations
1. Review security scan results above
2. Verify cost estimation matches budget
3. Check for destructive changes before applying
4. Ensure credentials are not hardcoded

### Next Steps
1. Apply changes: \`terraform apply tfplan-*.out\`
2. Monitor deployment: Check Cloud Eye dashboard
3. Verify services: Run health checks

EOF
    
    print_status $GREEN "SUCCESS" "Validation report generated: $REPORT_FILE"
}

# Main function
main() {
    echo "==================================="
    echo "PulseExpends Terraform Validation"
    echo "==================================="
    echo ""
    
    # Track overall status
    OVERALL_STATUS=0
    
    # Install/check required tools
    install_tools
    
    # Run validations
    validate_syntax || OVERALL_STATUS=1
    run_security_scans || OVERALL_STATUS=1
    estimate_costs
    validate_resource_limits
    
    # Generate plan
    generate_plan
    check_destructive_changes
    
    # Generate report
    generate_report
    
    echo ""
    if [ $OVERALL_STATUS -eq 0 ]; then
        print_status $GREEN "VALIDATION COMPLETE" "Infrastructure validation passed"
        print_status $GREEN "NEXT STEP" "Review the plan and apply with: terraform apply tfplan-*.out"
    else
        print_status $RED "VALIDATION FAILED" "Please fix the issues above before deployment"
        exit 1
    fi
}

# Run main function
main "$@"