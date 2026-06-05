#!/bin/bash

# PulseExpends Git Synchronization Script
# Automatically commits and pushes Terraform state changes

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

# Function to check if we're in a git repository
check_git_repo() {
    if [ ! -d ".git" ]; then
        print_status $RED "ERROR" "Not a git repository"
        exit 1
    fi
}

# Function to configure git identity
configure_git() {
    print_status $BLUE "INFO" "Configuring git identity..."
    
    # Check if git user is configured
    if [ -z "$(git config user.name)" ] || [ -z "$(git config user.email)" ]; then
        print_status $YELLOW "WARNING" "Git user identity not configured"
        
        # Try to get from environment or use defaults
        GIT_USER_NAME="${GIT_USER_NAME:-GitHub Actions}"
        GIT_USER_EMAIL="${GIT_USER_EMAIL:-actions@github.com}"
        
        git config user.name "$GIT_USER_NAME"
        git config user.email "$GIT_USER_EMAIL"
        
        print_status $GREEN "SUCCESS" "Git identity configured: $GIT_USER_NAME <$GIT_USER_EMAIL>"
    else
        print_status $GREEN "SUCCESS" "Git identity already configured: $(git config user.name) <$(git config user.email)>"
    fi
}

# Function to check for Terraform state changes
check_state_changes() {
    print_status $BLUE "INFO" "Checking for Terraform state changes..."
    
    # Check if terraform.tfstate exists
    if [ ! -f "terraform.tfstate" ]; then
        print_status $RED "ERROR" "terraform.tfstate not found"
        exit 1
    fi
    
    # Check if state file is tracked
    if ! git ls-files --error-unmatch terraform.tfstate >/dev/null 2>&1; then
        print_status $YELLOW "WARNING" "terraform.tfstate is not tracked in git"
        print_status $YELLOW "INFO" "Adding terraform.tfstate to git..."
        git add terraform.tfstate
    fi
    
    # Check for changes
    if git diff --quiet terraform.tfstate; then
        print_status $GREEN "INFO" "No changes in terraform.tfstate"
        STATE_CHANGED=false
    else
        print_status $YELLOW "INFO" "Changes detected in terraform.tfstate"
        STATE_CHANGED=true
        
        # Show diff summary
        print_status $BLUE "INFO" "State changes summary:"
        git diff --stat terraform.tfstate
    fi
}

# Function to check for output changes
check_output_changes() {
    print_status $BLUE "INFO" "Checking for Terraform output changes..."
    
    # Generate current outputs
    if command_exists terraform; then
        terraform output -json > terraform-outputs.json 2>/dev/null || true
        
        if [ -f "terraform-outputs.json" ]; then
            if [ ! -f "terraform-outputs-previous.json" ]; then
                print_status $YELLOW "INFO" "No previous outputs found for comparison"
                OUTPUT_CHANGED=true
            elif ! diff -q terraform-outputs.json terraform-outputs-previous.json >/dev/null 2>&1; then
                print_status $YELLOW "INFO" "Output changes detected"
                OUTPUT_CHANGED=true
                
                # Show output diff
                print_status $BLUE "INFO" "Output changes:"
                diff terraform-outputs-previous.json terraform-outputs.json || true
            else
                print_status $GREEN "INFO" "No changes in outputs"
                OUTPUT_CHANGED=false
            fi
        fi
    else
        print_status $YELLOW "WARNING" "Terraform not found, skipping output comparison"
        OUTPUT_CHANGED=false
    fi
}

# Function to create backup
create_backup() {
    print_status $BLUE "INFO" "Creating backup of current state..."
    
    TIMESTAMP=$(date +%Y%m%d%H%M%S)
    BACKUP_DIR=".terraform-backups"
    
    # Create backup directory if it doesn't exist
    mkdir -p "$BACKUP_DIR"
    
    # Backup state file
    if [ -f "terraform.tfstate" ]; then
        cp terraform.tfstate "$BACKUP_DIR/terraform.tfstate.backup.$TIMESTAMP"
        print_status $GREEN "SUCCESS" "State backup created: $BACKUP_DIR/terraform.tfstate.backup.$TIMESTAMP"
    fi
    
    # Backup outputs if they exist
    if [ -f "terraform-outputs.json" ]; then
        cp terraform-outputs.json "$BACKUP_DIR/terraform-outputs.json.backup.$TIMESTAMP"
    fi
    
    # Keep only last 10 backups
    ls -t "$BACKUP_DIR/"*.backup.* 2>/dev/null | tail -n +11 | xargs -r rm
}

# Function to commit changes
commit_changes() {
    local commit_message="$1"
    
    print_status $BLUE "INFO" "Committing changes: $commit_message"
    
    # Add state file
    git add terraform.tfstate
    
    # Add outputs if they exist
    if [ -f "terraform-outputs.json" ]; then
        git add terraform-outputs.json
    fi
    
    # Commit
    if git commit -m "$commit_message"; then
        print_status $GREEN "SUCCESS" "Changes committed"
    else
        print_status $RED "ERROR" "Commit failed"
        return 1
    fi
}

# Function to push changes
push_changes() {
    print_status $BLUE "INFO" "Pushing changes to remote..."
    
    # Get current branch
    CURRENT_BRANCH=$(git branch --show-current)
    
    # Push to remote
    if git push origin "$CURRENT_BRANCH"; then
        print_status $GREEN "SUCCESS" "Changes pushed to $CURRENT_BRANCH"
    else
        print_status $RED "ERROR" "Push failed"
        return 1
    fi
}

# Function to create deployment tag
create_deployment_tag() {
    print_status $BLUE "INFO" "Creating deployment tag..."
    
    DEPLOYMENT_ID=$(date +%Y%m%d%H%M%S)
    TAG_NAME="deploy-$DEPLOYMENT_ID"
    
    if git tag -a "$TAG_NAME" -m "Deployment $DEPLOYMENT_ID"; then
        print_status $GREEN "SUCCESS" "Created tag: $TAG_NAME"
        
        # Push tag
        if git push origin "$TAG_NAME"; then
            print_status $GREEN "SUCCESS" "Tag pushed to remote"
        else
            print_status $YELLOW "WARNING" "Failed to push tag to remote"
        fi
    else
        print_status $YELLOW "WARNING" "Failed to create tag"
    fi
}

# Function to generate deployment summary
generate_summary() {
    print_status $BLUE "INFO" "Generating deployment summary..."
    
    SUMMARY_FILE="deployment-summary-$(date +%Y%m%d%H%M%S).md"
    
    # Get git info
    COMMIT_HASH=$(git rev-parse --short HEAD)
    COMMIT_MESSAGE=$(git log -1 --pretty=%B)
    BRANCH=$(git branch --show-current)
    
    # Get Terraform info if available
    if command_exists terraform; then
        TERRAFORM_VERSION=$(terraform version | head -1)
        OUTPUTS=$(terraform output -json 2>/dev/null || echo "{}")
    else
        TERRAFORM_VERSION="Not available"
        OUTPUTS="{}"
    fi
    
    cat > "$SUMMARY_FILE" << EOF
# Deployment Summary
## PulseExpends Infrastructure - $(date)

### Deployment Information
- **Timestamp**: $(date)
- **Commit**: $COMMIT_HASH
- **Branch**: $BRANCH
- **Commit Message**: $COMMIT_MESSAGE
- **Terraform Version**: $TERRAFORM_VERSION

### Changes
- **State File Changed**: $STATE_CHANGED
- **Outputs Changed**: $OUTPUT_CHANGED

### Resources Deployed
$(if [ -f "terraform-outputs.json" ]; then
  echo "```json"
  cat terraform-outputs.json | head -50
  echo "```"
else
  echo "Outputs not available"
fi)

### Next Steps
1. Verify deployment: Check Cloud Eye dashboard
2. Test services: Run health checks
3. Monitor: Set up alerts for critical metrics
4. Documentation: Update deployment documentation

### Rollback Information
To rollback to previous state:
\`\`\`bash
# Checkout previous commit
git checkout HEAD~1

# Apply previous state
terraform apply
\`\`\`

### Backup Location
Backup created in: .terraform-backups/

EOF
    
    print_status $GREEN "SUCCESS" "Deployment summary generated: $SUMMARY_FILE"
    
    # Add summary to git
    git add "$SUMMARY_FILE"
    git commit -m "docs: Add deployment summary $(date +%Y%m%d%H%M%S)"
    git push origin "$BRANCH"
}

# Function to clean up old backups
cleanup_backups() {
    print_status $BLUE "INFO" "Cleaning up old backups..."
    
    BACKUP_DIR=".terraform-backups"
    
    if [ -d "$BACKUP_DIR" ]; then
        # Keep only last 30 days of backups
        find "$BACKUP_DIR" -name "*.backup.*" -mtime +30 -delete 2>/dev/null || true
        
        # Count remaining backups
        BACKUP_COUNT=$(find "$BACKUP_DIR" -name "*.backup.*" | wc -l)
        print_status $GREEN "SUCCESS" "Backup cleanup complete. $BACKUP_COUNT backups remaining."
    fi
}

# Main function
main() {
    echo "==================================="
    echo "PulseExpends Git Synchronization"
    echo "==================================="
    echo ""
    
    # Default commit message
    COMMIT_MESSAGE="terraform: Update state after deployment $(date +"%Y-%m-%d %H:%M:%S")"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -m|--message)
                COMMIT_MESSAGE="$2"
                shift 2
                ;;
            --no-push)
                NO_PUSH=true
                shift
                ;;
            --no-tag)
                NO_TAG=true
                shift
                ;;
            --summary-only)
                SUMMARY_ONLY=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  -m, --message TEXT    Custom commit message"
                echo "  --no-push             Don't push to remote"
                echo "  --no-tag              Don't create deployment tag"
                echo "  --summary-only        Only generate deployment summary"
                echo "  -h, --help           Show this help"
                exit 0
                ;;
            *)
                print_status $RED "ERROR" "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check if we're in a git repo
    check_git_repo
    
    # Configure git
    configure_git
    
    # Check for changes
    check_state_changes
    check_output_changes
    
    # If summary only mode
    if [ "$SUMMARY_ONLY" = true ]; then
        generate_summary
        exit 0
    fi
    
    # If no changes, exit
    if [ "$STATE_CHANGED" = false ] && [ "$OUTPUT_CHANGED" = false ]; then
        print_status $GREEN "INFO" "No changes to commit"
        exit 0
    fi
    
    # Create backup
    create_backup
    
    # Commit changes
    commit_changes "$COMMIT_MESSAGE"
    
    # Push changes (unless disabled)
    if [ "$NO_PUSH" != true ]; then
        push_changes
    fi
    
    # Create tag (unless disabled)
    if [ "$NO_TAG" != true ]; then
        create_deployment_tag
    fi
    
    # Generate summary
    generate_summary
    
    # Cleanup old backups
    cleanup_backups
    
    echo ""
    print_status $GREEN "SYNCHRONIZATION COMPLETE" "Git synchronization completed successfully"
    
    # Show next steps
    echo ""
    print_status $BLUE "NEXT STEPS" ""
    echo "1. Verify deployment: Check Cloud Eye dashboard"
    echo "2. Test endpoints: Run health checks"
    echo "3. Update documentation: Review deployment summary"
    echo "4. Monitor: Set up alerts for critical metrics"
}

# Run main function
main "$@"