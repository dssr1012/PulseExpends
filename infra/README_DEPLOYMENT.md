# PulseExpends Infrastructure Deployment Guide

## Overview
This Terraform configuration deploys the PulseExpends application infrastructure on Huawei Cloud with Enterprise Project support.

## Prerequisites

### 1. Huawei Cloud Account
- Huawei Cloud account with Enterprise Project enabled
- Access Key and Secret Key with sufficient permissions
- Project ID for the target region (la-south-2)

### 2. Local Tools
- Terraform >= 1.5.0
- Huawei Cloud CLI (optional)

### 3. Domain Configuration
- DuckDNS account for domain `pulseexpends.duckdns.org`
- Access to DuckDNS control panel to update DNS records

## Configuration

### 1. Update Credentials
Edit `terraform.tfvars` and update the following:

```hcl
# Huawei Cloud Credentials
access_key = "HPUAELE34ORKBY58ROT4"
secret_key = "Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439"
project_id = "your-actual-project-id-here"  # ← UPDATE THIS
```

### 2. Get Your Project ID
To find your project ID:
1. Log in to [Huawei Cloud Console](https://console.huaweicloud.com/)
2. Go to **IAM > Projects**
3. Find your project in region `la-south-2` (Santiago, Chile)
4. Copy the **Project ID**

### 3. Verify Enterprise Project
Ensure your account has Enterprise Project enabled. The configuration will create:
- Enterprise Project: `pulse-expendss` (type: `prod`)

### 4. Key Pair Setup
You need an SSH key pair named `pulse-expends-key`. Create it via:
```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 4096 -f pulse-expends-key -N ""
```

Upload the public key to Huawei Cloud:
1. Go to **ECS > Key Pairs**
2. Click **Import Key Pair**
3. Name: `pulse-expends-key`
4. Paste the public key content

## Deployment Steps

### Step 1: Initialize Terraform
```bash
cd /root/PulseExpends-Infra
./init-terraform.sh
```

### Step 2: Review the Plan
```bash
terraform show tfplan
```

### Step 3: Apply the Infrastructure
```bash
terraform apply tfplan
```

Type `yes` when prompted to confirm.

### Step 4: Get Deployment Outputs
After successful deployment, Terraform will output:
- **Enterprise Project ID**: ID of the created Enterprise Project
- **Domain IP Address**: Public IP address for DuckDNS configuration
- **Application URL**: URL to access the application
- **SSH Access Command**: Command to SSH into the ECS instance

### Step 5: Configure DuckDNS
1. Log in to [DuckDNS](https://www.duckdns.org/)
2. Go to your domain `pulseexpends.duckdns.org`
3. Update the A record with the **Domain IP Address** from Terraform outputs
4. Save the changes

### Step 6: Deploy Application
SSH into the ECS instance and deploy the application:
```bash
# Use the SSH command from Terraform outputs
ssh -i pulse-expends-key.pem root@<ECS_PUBLIC_IP>

# On the ECS instance, the application should auto-deploy via user-data script
# Check deployment status
cd /opt/pulse-expends
docker-compose ps
```

## Architecture

### Resources Created
1. **Enterprise Project**: `pulse-expendss` (production type)
2. **VPC**: 10.0.0.0/16 with public/private subnets
3. **EIP**: Elastic IP for domain `pulseexpends.duckdns.org`
4. **ECS Instance**: `c6.large.2` (2 vCPUs, 4GB RAM)
5. **Security Groups**: Open ports 22, 80, 443, 8080, 8000
6. **OBS Buckets**: 
   - `pulse-expends-data-dev` (encrypted, versioned)
   - `pulse-expends-documents-dev` (encrypted, versioned)
7. **KMS Key**: `alias/pulse-expends` for encryption
8. **Monitoring**: Cloud Eye alarms for CPU, memory, disk
9. **Anti-DDoS**: Basic DDoS protection

### Network Configuration
- **VPC CIDR**: 10.0.0.0/16
- **Public Subnets**: 10.0.1.0/24, 10.0.2.0/24
- **Private Subnets**: 10.0.101.0/24, 10.0.102.0/24
- **NAT Gateway**: Enabled for private subnet internet access

### Security
- All resources associated with Enterprise Project
- OBS buckets encrypted with KMS
- Security groups restrict access to necessary ports
- Anti-DDoS protection enabled
- No public database/redis in Phase 1

## Application Deployment

The ECS instance automatically deploys:
1. **MCP Server (Go)**: Port 8080
2. **PDF Parser Service (Python)**: Port 8000
3. **Nginx Reverse Proxy**: Ports 80/443

### Access URLs
- **Application**: `http://pulseexpends.duckdns.org`
- **MCP Server**: `http://pulseexpends.duckdns.org:8080`
- **PDF Parser**: `http://pulseexpends.duckdns.org:8000`
- **SSH Access**: `ssh -i pulse-expends-key.pem root@<ECS_PUBLIC_IP>`

## Phase 2 (Future)
The infrastructure is ready for Phase 2 expansion:
- **Database**: PostgreSQL RDS (currently disabled)
- **Redis**: Cache layer (currently disabled)
- **Load Balancer**: ELB (currently disabled)
- **Auto Scaling**: AS group (currently disabled)
- **SSL/TLS**: Certificate management (currently disabled)

## Monitoring

### Cloud Eye Dashboard
Access via Huawei Cloud Console:
1. Go to **Cloud Eye > Dashboard**
2. Find dashboard: `pulseexpends-dev-<id>-dashboard`

### Alarms Configured
- CPU Usage > 80%
- Memory Usage > 85%
- Disk Usage > 90%

## Cost Estimation

### Monthly Costs (Approximate)
- **ECS c6.large.2**: ~$50-70/month
- **EIP (5Mbps)**: ~$10-15/month
- **OBS Storage (50GB)**: ~$1-2/month
- **NAT Gateway**: ~$20-30/month
- **Total**: ~$80-120/month

### Cost Optimization
- Use reserved instances for long-term savings
- Monitor and adjust EIP bandwidth as needed
- Implement auto-scaling for production loads

## Troubleshooting

### Common Issues

1. **Terraform Init Fails**
   - Check credentials in `terraform.tfvars`
   - Verify project_id is correct
   - Ensure region `la-south-2` is accessible

2. **SSH Access Denied**
   - Verify key pair exists in Huawei Cloud
   - Check security group allows port 22
   - Ensure EIP is properly associated

3. **Application Not Accessible**
   - Check ECS instance status
   - Verify Docker containers are running
   - Check security group rules
   - Confirm DuckDNS propagation (can take 5-10 minutes)

4. **Enterprise Project Issues**
   - Verify account has Enterprise Project permissions
   - Check quota limits
   - Contact Huawei Cloud support if needed

### Useful Commands
```bash
# Check Terraform state
terraform state list

# View outputs
terraform output

# Destroy infrastructure (careful!)
terraform destroy

# Update DuckDNS manually (if needed)
curl "https://www.duckdns.org/update?domains=pulseexpends&token=YOUR_TOKEN&ip=YOUR_IP"
```

## Support

For issues:
1. Check Huawei Cloud Service Status
2. Review Terraform error messages
3. Check Cloud Eye logs
4. Contact Huawei Cloud Support

## Next Steps After Deployment

1. **Test Application**: Access `http://pulseexpends.duckdns.org`
2. **Configure Monitoring**: Set up additional Cloud Eye alarms
3. **Backup Strategy**: Configure OBS lifecycle policies
4. **Security Hardening**: Review security group rules
5. **Performance Testing**: Load test the application

## Notes
- All resources are tagged with: Project=PulseExpends, Environment=dev, ManagedBy=Terraform
- Enterprise Project ensures resource isolation and billing separation
- DuckDNS provides free dynamic DNS for development
- Phase 2 components (DB, Redis, LB) can be enabled by setting variables to `true`