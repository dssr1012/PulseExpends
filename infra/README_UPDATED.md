# PulseExpends Infrastructure - Huawei Cloud

Infrastructure as Code for PulseExpends application deployment on Huawei Cloud with Enterprise Project and DuckDNS domain support.

## 🚀 Quick Start

### Prerequisites

1. **Huawei Cloud Account** with:
   - Access Key and Secret Key
   - Project ID
   - Enterprise Project enabled

2. **Terraform** v1.5.0 or later
   ```bash
   # Install Terraform
   curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
   sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
   sudo apt-get update && sudo apt-get install terraform
   ```

3. **SSH Key Pair** for ECS instances
   ```bash
   ssh-keygen -t rsa -b 4096 -f pulse-expends-key.pem
   ```

### Credentials Setup

You have received the following credentials:

```
Access Key ID: HPUAELE34ORKBY58ROT4
Secret Access Key: Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439
Project ID: [Your Project ID]
```

#### Option 1: Environment Variables (Recommended)
```bash
export HUAWEICLOUD_ACCESS_KEY="HPUAELE34ORKBY58ROT4"
export HUAWEICLOUD_SECRET_KEY="Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439"
export HUAWEICLOUD_PROJECT_ID="your-project-id-here"
```

#### Option 2: terraform.tfvars File
Copy the example configuration and update with your credentials:
```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
```

### Domain Configuration

The project is configured to use DuckDNS domain: **pulseexpends.duckdns.org**

1. **Get DuckDNS Token**:
   - Visit https://www.duckdns.org
   - Login with your preferred method (GitHub, Google, etc.)
   - Go to your domains
   - Copy your token

2. **Set DuckDNS Token**:
   ```bash
   export DUCKDNS_TOKEN="your-duckdns-token-here"
   ```

## 📁 Project Structure

```
PulseExpends-Infra/
├── main_updated.tf              # Main Terraform configuration (updated)
├── variables.tf                 # Variable definitions (updated)
├── outputs.tf                   # Output values
├── terraform.tfvars.example     # Example configuration
├── terraform.tfvars             # Your configuration (create this)
├── deploy.sh                    # Deployment automation script
├── scripts/
│   └── configure-duckdns.sh     # DuckDNS configuration script
└── environments/                # Environment-specific configurations
    ├── dev/                     # Development environment
    ├── staging/                 # Staging environment
    └── prod/                    # Production environment
```

## 🏗️ Architecture

### Phase 1 (Current Deployment - 1-1000 users)
- **Enterprise Project**: `pulse-expendss`
- **Compute**: ECS instances (c6.large.2 - 2vCPU, 4GB RAM)
- **Storage**: Huawei Cloud OBS (Object Storage Service)
- **Network**: VPC with public/private subnets, NAT Gateway
- **Domain**: pulseexpends.duckdns.org
- **Security**: Security Groups, KMS encryption

### Phase 2 (100-10,000 users)
- **Database**: PostgreSQL RDS
- **Cache**: Redis
- **Load Balancer**: ELB
- **Auto Scaling**: AS groups

### Phase 3 (10,000+ users)
- **Kubernetes**: CCE (Cloud Container Engine)
- **Multi-AZ**: High availability
- **CDN**: Content Delivery Network
- **WAF**: Web Application Firewall

## 🚢 Deployment

### 1. Initialize Terraform
```bash
# Make deployment script executable
chmod +x deploy.sh

# Initialize Terraform
terraform init
```

### 2. Review Deployment Plan
```bash
# Review what will be created
./deploy.sh --action plan --environment dev
```

### 3. Deploy Infrastructure
```bash
# Deploy with automatic DuckDNS configuration
./deploy.sh --action deploy --environment dev --auto-confirm

# Or deploy step by step
./deploy.sh --action deploy --environment dev
```

### 4. Configure DuckDNS (if not automatic)
```bash
# Get the public IP from Terraform outputs
terraform output domain_ip_address

# Update DuckDNS manually
curl "https://www.duckdns.org/update?domains=pulseexpends&token=YOUR_TOKEN&ip=YOUR_IP"
```

### 5. Verify Deployment
```bash
# Get deployment outputs
./deploy.sh --action outputs --environment dev

# Check application health
curl http://pulseexpends.duckdns.org/health
curl http://pulseexpends.duckdns.org:8080/health
curl http://pulseexpends.duckdns.org:8000/health
```

## 🔧 Configuration

### Environment Variables
```bash
# Required
export HUAWEICLOUD_ACCESS_KEY="HPUAELE34ORKBY58ROT4"
export HUAWEICLOUD_SECRET_KEY="Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439"
export HUAWEICLOUD_PROJECT_ID="your-project-id"

# Optional (for DuckDNS)
export DUCKDNS_TOKEN="your-duckdns-token"
```

### terraform.tfvars Configuration
```hcl
# Huawei Cloud Credentials
access_key = "HPUAELE34ORKBY58ROT4"
secret_key = "Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439"
project_id = "your-project-id-here"

# Region
region = "la-south-2"  # Santiago, Chile

# Environment
environment = "dev"
project_name = "pulseexpends"

# Enterprise Project
enable_enterprise_project = true
enterprise_project_name = "pulse-expendss"
enterprise_project_type = "prod"

# Domain Configuration
enable_domain = true
domain_name = "pulseexpends.duckdns.org"
create_dns_record = false  # Set to true if using Huawei Cloud DNS
domain_bandwidth_size = 5  # Mbps

# ECS Configuration
ecs_instance_type = "c6.large.2"
ecs_key_pair = "pulse-expends-key"
ecs_instance_count = 1

# OBS Configuration
obs_buckets = {
  main = {
    name          = "pulse-expends-data-dev"
    storage_class = "STANDARD"
    versioning    = true
    encryption    = true
  }
}
```

## 📊 Outputs

After deployment, you'll get:

```bash
# Get all outputs
terraform output

# Get specific outputs
terraform output application_url
terraform output domain_ip_address
terraform output ssh_access_command
```

### Expected Outputs:
- **Enterprise Project ID**: The created Enterprise Project
- **VPC ID**: Virtual Private Cloud ID
- **ECS Public IPs**: Public IP addresses of instances
- **Domain IP**: Public IP for the domain
- **Application URLs**: Access URLs for services
- **SSH Access**: Command to access instances
- **OBS Buckets**: Created storage buckets

## 🔒 Security

### Created Resources:
1. **Enterprise Project**: Isolated environment for all resources
2. **VPC with Security Groups**: Network isolation and firewall rules
3. **KMS Encryption**: Server-side encryption for OBS buckets
4. **IAM Policies**: Least privilege access control
5. **Anti-DDoS**: Basic DDoS protection enabled

### Security Best Practices:
- All OBS buckets have encryption enabled
- Security groups restrict access to necessary ports only
- SSH key-based authentication for ECS instances
- Regular security updates via user-data script
- Monitoring and alerting configured

## 🚨 Monitoring

### Cloud Eye Dashboard
- CPU/Memory usage monitoring
- Network traffic monitoring
- Disk I/O monitoring
- Custom alarms for critical metrics

### Access URLs:
- **Application**: http://pulseexpends.duckdns.org
- **MCP Server**: http://pulseexpends.duckdns.org:8080
- **Python Service**: http://pulseexpends.duckdns.org:8000
- **Cloud Eye**: Access via Huawei Cloud Console

## 🛠️ Management

### SSH Access
```bash
# Use the SSH command from outputs
ssh -i pulse-expends-key.pem root@<ecs-public-ip>
```

### Application Management
```bash
# Deploy application
cd /opt/pulse-expends
docker-compose up -d

# View logs
docker-compose logs -f

# Restart services
docker-compose restart

# Backup data
/opt/backup/backup.sh
```

### Terraform Management
```bash
# View state
terraform show

# Modify and re-apply
terraform apply

# Destroy everything (CAUTION!)
terraform destroy
```

## 🔄 Updates and Maintenance

### Update Application
1. Push new code to repository
2. SSH into ECS instance
3. Pull latest changes
4. Restart Docker containers

### Scale Resources
1. Update `ecs_instance_count` in terraform.tfvars
2. Run `terraform apply`

### Update Domain
1. Update `domain_name` in terraform.tfvars
2. Run `terraform apply`
3. Update DuckDNS configuration

## 🧪 Testing

### Run All Tests
```bash
cd ../PulseExpends
./run_tests.sh
```

### Test Specific Components
```bash
# Go tests (MCP Server)
./run_tests.sh go

# Python tests (PDF Parser)
./run_tests.sh python

# Security tests
./run_tests.sh security

# Integration tests
./run_tests.sh integration

# Performance tests
./run_tests.sh performance
```

### Test Coverage Reports
- Go coverage: `coverage.html`
- Python coverage: `coverage_html/`
- Security report: `security_report.json`

## 🗑️ Cleanup

### Destroy Infrastructure
```bash
# Destroy with confirmation
./deploy.sh --action destroy --environment dev

# Destroy without confirmation (CAUTION!)
./deploy.sh --destroy --environment dev --auto-confirm
```

### Cleanup Steps:
1. Destroy Terraform resources
2. Remove DuckDNS entry (optional)
3. Delete local Terraform state
4. Remove SSH key pair

## 🆘 Troubleshooting

### Common Issues:

1. **Credentials Error**
   ```
   Error: error creating IAM client: Incorrect IAM authentication information
   ```
   **Solution**: Verify access_key, secret_key, and project_id

2. **Enterprise Project Error**
   ```
   Error: error creating enterprise project
   ```
   **Solution**: Check if Enterprise Project feature is enabled in your account

3. **Domain Configuration Error**
   ```
   Error: error creating EIP
   ```
   **Solution**: Check quota limits for EIP in your region

4. **SSH Connection Failed**
   ```
   Permission denied (publickey)
   ```
   **Solution**: Ensure SSH key pair exists and is correctly referenced

5. **DuckDNS Update Failed**
   ```
   Response: KO
   ```
   **Solution**: Verify DuckDNS token and domain ownership

### Debug Commands:
```bash
# Check Terraform version
terraform version

# Validate configuration
terraform validate

# Show execution plan
terraform plan

# Show current state
terraform state list

# Get specific resource info
terraform state show <resource>
```

## 📞 Support

### Huawei Cloud Support
- **Region**: la-south-2 (Santiago, Chile)
- **Console**: https://console-intl.huaweicloud.com
- **Documentation**: https://support.huaweicloud.com/intl/en-us/

### DuckDNS Support
- **Website**: https://www.duckdns.org
- **Documentation**: https://www.duckdns.org/spec.jsp

### Project Documentation
- **Repository**: https://github.com/dssr1012/PulseExpends
- **Infrastructure**: https://github.com/dssr1012/PulseExpends-Infra

## 📝 Notes

### Important Considerations:
1. **Cost Management**: Monitor costs in Huawei Cloud Console
2. **Backup Strategy**: Regular backups of OBS data
3. **Security Updates**: Keep ECS instances updated
4. **Monitoring**: Set up alerts for critical metrics
5. **Scaling**: Adjust instance types based on load

### Next Steps After Deployment:
1. ✅ Configure DuckDNS domain
2. ✅ Test application endpoints
3. ✅ Set up monitoring alerts
4. ✅ Configure backup schedule
5. ✅ Implement CI/CD pipeline
6. ✅ Set up logging aggregation
7. ✅ Configure SSL certificates (optional)

### Production Checklist:
- [ ] Enable WAF (Web Application Firewall)
- [ ] Configure Auto Scaling
- [ ] Set up database (Phase 2)
- [ ] Implement Redis caching (Phase 2)
- [ ] Configure CDN (Phase 3)
- [ ] Set up disaster recovery
- [ ] Implement comprehensive monitoring
- [ ] Configure automated backups

## 📄 License

This infrastructure code is part of the PulseExpends project. See the main repository for license information.

## 🙏 Acknowledgments

- Huawei Cloud for infrastructure services
- DuckDNS for free dynamic DNS
- HashiCorp for Terraform
- Open source community for tools and libraries

---

**Last Updated**: $(date)
**Terraform Version**: >= 1.5.0
**Huawei Cloud Region**: la-south-2
**Domain**: pulseexpends.duckdns.org
**Enterprise Project**: pulse-expendss