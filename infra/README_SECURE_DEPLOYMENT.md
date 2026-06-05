# PulseExpends Infrastructure Deployment Package
# Optimized for 10 concurrent users in Santiago region (la-south-2)

## Overview
This package provides a complete, secure infrastructure deployment for PulseExpends on Huawei Cloud, optimized for 10 concurrent users with cost-effective resource sizing and security best practices.

## Directory Structure
```
infra/
├── terraform/                    # Terraform configuration
│   ├── main.tf                  # Main infrastructure
│   ├── variables.tf             # Variable definitions
│   ├── rds.tf                   # Database configuration
│   ├── outputs.tf              # Output values
│   ├── security.tf             # Enhanced security rules
│   └── monitoring.tf           # Cloud Eye monitoring
├── scripts/                     # Deployment scripts
│   ├── deploy.sh              # Main deployment script
│   ├── secure-credentials.sh  # Credential management
│   ├── validate-plan.sh       # Terraform validation
│   └── git-sync.sh           # Git synchronization
├── .gitignore                 # Git ignore rules
├── terraform.tfvars.example   # Configuration template
└── README_SECURE_DEPLOYMENT.md # Secure deployment guide
```

## 1. Enhanced Security Configuration

### Secure Credential Management
**DO NOT** commit credentials to version control. Use one of these methods:

#### Option A: Environment Variables (Recommended)
```bash
export TF_VAR_access_key="your_access_key"
export TF_VAR_secret_key="your_secret_key"
export TF_VAR_project_id="your_project_id"
export TF_VAR_rds_password="$(openssl rand -base64 32)"
```

#### Option B: Encrypted Secrets File
```bash
# Create encrypted secrets
gpg --symmetric --cipher-algo AES256 terraform.secrets.tfvars

# Decrypt during deployment
gpg --decrypt terraform.secrets.tfvars.gpg > terraform.tfvars
```

#### Option C: Huawei Cloud IAM Agency
```terraform
provider "huaweicloud" {
  region     = var.region
  access_key = var.access_key
  secret_key = var.secret_key
  project_id = var.project_id
  
  # Use agency for enhanced security
  agency_name        = var.agency_name
  agency_domain_name = var.agency_domain_name
}
```

## 2. Infrastructure Resources (10 Users)

### Compute (ECS)
- **Instance Type**: `s6.large.2` (2 vCPUs, 4GB RAM)
- **Count**: 1 instance (no auto-scaling needed)
- **OS**: Ubuntu 22.04 LTS
- **Storage**: 40GB SSD system disk
- **Cost**: ~$30/month

### Database (RDS PostgreSQL)
- **Instance Type**: `rds.pg.n1.large.2` (2 vCPUs, 4GB RAM)
- **Storage**: 50GB SSD
- **Backup**: 7-day retention, daily at 02:00 UTC
- **No HA**: Single instance for cost optimization
- **Cost**: ~$50/month

### Networking
- **VPC**: `10.0.0.0/16`
- **Public Subnet**: `10.0.1.0/24`
- **Security Group**: Restricted access rules
- **Elastic IP**: 1 static IP

### Storage (OBS)
- **Bucket**: `pulseexpends-documents-dev`
- **Storage Class**: STANDARD
- **Versioning**: Enabled
- **Encryption**: Server-side (KMS optional)
- **Cost**: ~$5/month (100GB)

### Monitoring (Cloud Eye)
- **CPU Alert**: >80% for 5 minutes
- **Memory Alert**: >85% for 5 minutes
- **Disk Alert**: >90% usage
- **RDS Connections**: >50 concurrent
- **Cost**: Free tier

## 3. Security Hardening

### Network Security
1. **Restrict SSH Access**: Only from specific IP ranges
2. **Close Unnecessary Ports**: Only 22, 80, 443, 8080, 8000
3. **VPC Flow Logs**: Enabled for audit trail
4. **Security Group Rules**: Least privilege principle

### Data Security
1. **RDS Encryption**: At-rest encryption enabled
2. **OBS Encryption**: Server-side encryption
3. **KMS Integration**: Optional for enhanced security
4. **Password Management**: Auto-generated strong passwords

### Access Control
1. **IAM Policies**: Minimum required permissions
2. **Agency Usage**: For cross-account access
3. **Credential Rotation**: 90-day automated rotation
4. **Audit Logging**: All API calls logged

## 4. Deployment Workflow

### Phase 1: Preparation
```bash
# 1. Clone repository
git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends/infra

# 2. Set up credentials securely
./scripts/secure-credentials.sh

# 3. Review configuration
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
```

### Phase 2: Validation
```bash
# 1. Initialize Terraform
terraform init

# 2. Validate configuration
./scripts/validate-plan.sh

# 3. Check security
tfsec .
checkov --directory .

# 4. Estimate costs
infracost breakdown --path .
```

### Phase 3: Deployment
```bash
# 1. Generate execution plan
terraform plan -out=tfplan

# 2. Review plan (human verification required)
terraform show tfplan | grep -E "(Plan:|~|\+|\-)"

# 3. Apply configuration
terraform apply tfplan

# 4. Save outputs
terraform output -json > deployment-outputs.json
```

### Phase 4: Post-Deployment
```bash
# 1. Sync state to Git
./scripts/git-sync.sh

# 2. Run health checks
curl http://$(terraform output -raw ecs_public_ip)/health

# 3. Verify services
./scripts/verify-deployment.sh
```

## 5. CI/CD Integration

### GitHub Actions Workflow
```yaml
name: Terraform Deployment
on:
  push:
    branches: [main]
    paths:
      - 'infra/**'
  pull_request:
    branches: [main]
    paths:
      - 'infra/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        
      - name: Terraform Validate
        run: terraform validate
        
      - name: Terraform Security Scan
        uses: bridgecrewio/checkov-action@master
        
      - name: Terraform Cost Estimation
        uses: infracost/infracost-action@v1

  deploy:
    needs: validate
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        
      - name: Terraform Apply
        run: terraform apply -auto-approve
        env:
          TF_VAR_access_key: ${{ secrets.HUAWEICLOUD_ACCESS_KEY }}
          TF_VAR_secret_key: ${{ secrets.HUAWEICLOUD_SECRET_KEY }}
          TF_VAR_project_id: ${{ secrets.HUAWEICLOUD_PROJECT_ID }}
          
      - name: Commit State Changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add terraform.tfstate
          git commit -m "terraform: Update state after deployment"
          git push
```

## 6. Cost Optimization

### Monthly Estimate: $80-120 USD
- **ECS**: $30-40
- **RDS**: $50-70  
- **OBS**: $2-5
- **EIP**: $5-10
- **Data Transfer**: $0-5

### Cost-Saving Measures
1. **Right-sizing**: s6.large.2 instead of c6.large.4
2. **No Auto-scaling**: Single instance for 10 users
3. **No Load Balancer**: Direct ECS access
4. **No Multi-AZ**: Single availability zone
5. **Reduced Storage**: 50GB instead of 100GB

## 7. Monitoring & Alerting

### Cloud Eye Alarms
```terraform
resource "huaweicloud_ces_alarmrule" "cpu_high" {
  alarm_name        = "high-cpu-usage"
  alarm_description = "CPU usage > 80% for 5 minutes"
  metric {
    namespace  = "SYS.ECS"
    metric_name = "cpu_util"
    dimensions {
      name  = "instance_id"
      value = huaweicloud_compute_instance.main.id
    }
  }
  condition {
    period              = 300
    filter             = "average"
    comparison_operator = ">"
    value              = 80
    unit               = "%"
    count              = 1
  }
  alarm_actions {
    type = "notification"
    notification_list = [var.monitoring_email]
  }
}
```

## 8. Disaster Recovery

### Backup Strategy
1. **RDS Automated Backups**: Daily with 7-day retention
2. **OBS Versioning**: All document changes tracked
3. **Terraform State**: Stored in Git with version history
4. **Configuration Backup**: All Terraform files in Git

### Recovery Procedures
1. **Infrastructure Recovery**: `terraform apply` from Git
2. **Database Recovery**: RDS point-in-time restore
3. **Data Recovery**: OBS version restoration
4. **Rollback Procedure**: Git revert + terraform apply

## 9. Security Compliance

### PCI-DSS Considerations
1. **Network Segmentation**: VPC with restricted subnets
2. **Encryption**: Data at rest and in transit
3. **Access Control**: IAM with least privilege
4. **Logging & Monitoring**: Cloud Eye + VPC Flow Logs
5. **Vulnerability Management**: Regular security updates

### GDPR Considerations
1. **Data Residency**: Santiago region (Chile)
2. **Data Protection**: Encryption enabled
3. **Access Logs**: All access attempts logged
4. **Data Portability**: Export capabilities via OBS
5. **Right to Erasure**: Automated data deletion procedures

## 10. Next Steps

### Immediate Actions
1. **Remove hardcoded credentials** from existing terraform.tfvars
2. **Implement IAM agency** with least privilege
3. **Update security group rules** to restrict access
4. **Enable monitoring alerts** for critical metrics
5. **Set up backup verification** procedures

### Future Enhancements
1. **Multi-region deployment** for higher availability
2. **Auto-scaling** for traffic spikes
3. **CDN integration** for global performance
4. **WAF protection** for web application security
5. **Advanced monitoring** with custom dashboards

## Support
For issues or questions:
1. Check deployment logs in `infra/deployment.log`
2. Review Terraform state in `terraform.tfstate`
3. Monitor Cloud Eye for resource metrics
4. Contact DevOps team for critical issues