# PulseExpends Infrastructure Deployment Guide
# Huawei Cloud Santiago Region (la-south-2) - 10 Concurrent Users

## Overview
This guide provides step-by-step instructions for deploying the PulseExpends infrastructure on Huawei Cloud using Terraform. The infrastructure is optimized for **10 concurrent users** with cost-effective resource sizing and security best practices.

## Prerequisites

### 1. Huawei Cloud Account
- Huawei Cloud account with billing enabled
- Access Key and Secret Key with appropriate permissions
- Project ID for Santiago region (la-south-2)

### 2. Local Tools
- **Terraform** >= 1.5.0
- **Git** for version control
- **OpenSSL** for generating secure passwords
- **jq** for JSON processing (optional)
- **tfsec** for security scanning (optional)
- **checkov** for security scanning (optional)
- **infracost** for cost estimation (optional)

### 3. Domain Configuration (Optional)
- DuckDNS account for dynamic DNS (if using custom domain)
- Or Huawei Cloud DNS zone (if using Huawei Cloud DNS)

## Quick Start

### Step 1: Clone and Setup
```bash
# Clone repository
git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends/infra

# Make scripts executable
chmod +x scripts/*.sh
```

### Step 2: Secure Credential Setup
```bash
# Method A: Environment Variables (Recommended)
export HUAWEICLOUD_ACCESS_KEY="your_access_key"
export HUAWEICLOUD_SECRET_KEY="your_secret_key"
export HUAWEICLOUD_PROJECT_ID="your_project_id"

# Generate secure configuration
./scripts/secure-credentials.sh generate

# Method B: Create terraform.tfvars manually
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
```

### Step 3: Validate Configuration
```bash
# Run comprehensive validation
./scripts/validate-plan.sh

# Expected output:
# [INFO] Validating Terraform syntax...
# [SUCCESS] Terraform configuration is valid
# [INFO] Running security scans...
# [INFO] Estimating infrastructure costs...
# [INFO] Generating Terraform execution plan...
```

### Step 4: Deploy Infrastructure
```bash
# Initialize Terraform
terraform init

# Review plan
terraform plan -out=tfplan

# Apply configuration
terraform apply tfplan

# Expected deployment time: 10-15 minutes
```

### Step 5: Verify Deployment
```bash
# Get deployment outputs
terraform output

# Test connectivity
./scripts/test-connectivity.sh

# Run health checks
./scripts/health-check.sh
```

### Step 6: Sync to Git
```bash
# Commit and push state changes
./scripts/git-sync.sh

# Or with custom message
./scripts/git-sync.sh -m "terraform: Initial deployment for 10-user infrastructure"
```

## Architecture Details

### Resource Sizing (10 Users)
| Resource | Type | Specification | Estimated Cost |
|----------|------|---------------|----------------|
| **ECS** | s6.large.2 | 2 vCPUs, 4GB RAM | ~$30/month |
| **RDS** | rds.pg.n1.large.2 | 2 vCPUs, 4GB RAM, 50GB SSD | ~$50/month |
| **Storage** | OBS Standard | 100GB | ~$5/month |
| **Networking** | EIP + VPC | 100Mbps bandwidth | ~$10/month |
| **Monitoring** | Cloud Eye | Free tier | $0/month |
| **Total** | | | **~$95/month** |

### Security Configuration
1. **Restricted Access**: Security groups limit access to specific IP ranges
2. **Encryption**: RDS and OBS encryption enabled
3. **IAM Least Privilege**: Minimum required permissions
4. **Network Isolation**: VPC with private subnets
5. **Monitoring**: Cloud Eye alerts for critical metrics

### High Availability
- **Single AZ deployment** for cost optimization
- **Automated backups**: RDS daily backups with 7-day retention
- **Manual failover**: Terraform scripts for disaster recovery

## Security Best Practices

### 1. Credential Management
```bash
# NEVER commit credentials to git
echo "terraform.tfvars" >> .gitignore
echo "*.tfstate*" >> .gitignore
echo ".terraform/" >> .gitignore

# Use environment variables
export TF_VAR_access_key="your_key"
export TF_VAR_secret_key="your_secret"
export TF_VAR_project_id="your_project"

# Or use encrypted secrets
gpg --symmetric --cipher-algo AES256 terraform.secrets.tfvars
```

### 2. Network Security
- SSH access restricted to specific IPs
- HTTP/HTTPS access limited to necessary ranges
- Database only accessible from ECS instances
- VPC Flow Logs enabled for audit trail

### 3. Data Protection
- RDS encryption at rest
- OBS server-side encryption
- Automated backups with retention policy
- Secure password generation

## Monitoring & Alerting

### Cloud Eye Alarms
The following alarms are configured:
1. **ECS CPU > 80%** for 5 minutes
2. **ECS Memory > 85%** for 5 minutes
3. **ECS Disk > 90%** usage
4. **RDS CPU > 70%** for 5 minutes
5. **RDS Connections > 50** for 5 minutes
6. **RDS Disk > 85%** usage

### Dashboard
Access the Cloud Eye dashboard:
```bash
terraform output monitoring_dashboard_url
```

## Cost Optimization

### Current Optimization
1. **Right-sized instances**: s6.large.2 for ECS, n1.large.2 for RDS
2. **Single instance**: No auto-scaling for 10 users
3. **No load balancer**: Direct ECS access
4. **Single AZ**: Reduced cross-AZ data transfer costs
5. **Optimized storage**: 50GB RDS, 100GB OBS

### Future Scaling
When user count increases:
1. **50 users**: Upgrade to s6.xlarge.4 (4 vCPU, 8GB RAM)
2. **100 users**: Add auto-scaling (2-4 instances)
3. **500+ users**: Implement load balancer + multi-AZ RDS

## Disaster Recovery

### Backup Strategy
1. **RDS**: Daily automated backups (7-day retention)
2. **OBS**: Versioning enabled for all objects
3. **Terraform State**: Git version control
4. **Configuration**: Git repository

### Recovery Procedures
```bash
# 1. Restore from backup
terraform apply

# 2. Verify deployment
terraform output
./scripts/health-check.sh

# 3. Test services
curl http://$(terraform output -raw ecs_public_ip)/health
```

## Troubleshooting

### Common Issues

#### 1. Credential Errors
```bash
# Error: Invalid access key or secret key
# Solution: Verify credentials and region
export HUAWEICLOUD_ACCESS_KEY="correct_key"
export HUAWEICLOUD_SECRET_KEY="correct_secret"
export HUAWEICLOUD_REGION="la-south-2"
```

#### 2. Resource Limits
```bash
# Error: Quota exceeded
# Solution: Request quota increase or optimize resources
# Contact Huawei Cloud support or reduce instance sizes
```

#### 3. Network Connectivity
```bash
# Error: Cannot connect to ECS
# Solution: Check security group rules
# Update allowed_ssh_ips in terraform.tfvars
```

#### 4. Terraform State Issues
```bash
# Error: State file corruption
# Solution: Restore from backup
cp .terraform-backups/terraform.tfstate.backup.* terraform.tfstate
terraform init
```

### Debug Commands
```bash
# Check Terraform version
terraform version

# Validate configuration
terraform validate

# Show current state
terraform show

# List resources
terraform state list

# Refresh state
terraform refresh

# Debug Terraform
TF_LOG=DEBUG terraform plan
```

## Maintenance

### Regular Tasks
1. **Weekly**: Check Cloud Eye alerts and logs
2. **Monthly**: Review costs and optimize resources
3. **Quarterly**: Rotate credentials and update security groups
4. **Bi-annually**: Test disaster recovery procedures

### Updates
```bash
# Update Terraform providers
terraform init -upgrade

# Update modules
terraform get -update

# Apply updates
terraform plan
terraform apply
```

## Support

### Getting Help
1. **Terraform Documentation**: https://developer.hashicorp.com/terraform
2. **Huawei Cloud Documentation**: https://support.huaweicloud.com/intl/en-us/
3. **GitHub Issues**: https://github.com/dssr1012/PulseExpends/issues

### Emergency Contacts
- **Infrastructure Team**: infrastructure@pulseexpends.com
- **Security Team**: security@pulseexpends.com
- **Huawei Cloud Support**: https://support.huaweicloud.com/intl/en-us/contact-us/

## Appendix

### Resource Naming Convention
- Format: `{project}-{environment}-{resource}-{identifier}`
- Example: `pulseexpends-dev-ecs-abc123`

### Tagging Strategy
All resources are tagged with:
- `Project`: pulseexpends
- `Environment`: dev/staging/prod
- `ManagedBy`: Terraform
- `CostCenter`: Engineering

### Cost Tracking
```bash
# Monthly cost estimation
infracost breakdown --path .

# Daily cost monitoring
# Set up Huawei Cloud Cost Center alerts
```

### Performance Benchmarks
- **Expected RPS**: 50-100 requests/second
- **Database Connections**: Max 50 concurrent
- **Response Time**: < 200ms for 95% of requests
- **Uptime SLA**: 99.5% (single AZ)

---

**Last Updated**: $(date +%Y-%m-%d)
**Terraform Version**: 1.5.0+
**Huawei Cloud Region**: la-south-2 (Santiago)
**Target Users**: 10 concurrent
**Monthly Budget**: $100 USD