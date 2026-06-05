# PulseExpends Infrastructure - Corrected Deployment Configuration
# =================================================================

## 🔧 CRITICAL CHANGES FROM PREVIOUS DEPLOYMENT

### 1. **ENTERPRISE PROJECT FIX** ✅
**Previous Issue**: Resources deployed in default enterprise project (ID: "0")
**Fix**: All resources will now be deployed in enterprise project "pulse-expends"

**Changes in terraform.tfvars:**
```hcl
# BEFORE (wrong):
enable_enterprise_project = false
enterprise_project_name = "pulse-expends"
enterprise_project_type = "prod"

# AFTER (correct):
enable_enterprise_project = true  # ← CRITICAL: Changed from false to true
enterprise_project_name = "pulse-expends"
enterprise_project_type = "prod"
```

**Impact on resources:**
- VPC: Will be created in "pulse-expends" enterprise project
- ECS: Will be created in "pulse-expends" enterprise project  
- RDS: Will be created in "pulse-expends" enterprise project
- Security Groups: Will be created in "pulse-expends" enterprise project

### 2. **RDS POSTGRESQL VERSION UPGRADE** ✅
**Previous Version**: PostgreSQL 15
**New Version**: PostgreSQL 17.9 (latest supported)

**Changes in terraform.tfvars:**
```hcl
# BEFORE:
rds_engine_version = "15"

# AFTER:
rds_engine_version = "17.9"  # ← Updated to latest PostgreSQL 17.9
```

**Benefits:**
- Latest security patches
- Performance improvements
- New features in PostgreSQL 17.x
- Better compatibility with modern applications

### 3. **IAM PERMISSION FIXES** ⚠️
**Missing permissions identified in previous deployment:**
1. `vpc:securityGroupRules:create` - For SSH rule creation
2. `kms:cmk:create` - For KMS key creation
3. `obs:bucket:CreateBucket` - For OBS storage
4. `rds:*` - For RDS PostgreSQL creation

**Recommendation**: Add these permissions to IAM user before deployment

### 4. **SECURITY GROUP FIX** 🔒
**Previous Issue**: SSH port (22) rule creation failed
**Root Cause**: IAM permission `vpc:securityGroupRules:create` missing

**Fix**: Ensure IAM user has proper permissions before deployment

## 📋 CORRECTED DEPLOYMENT PLAN

### Resources to be Created in Enterprise Project "pulse-expends":

1. **Enterprise Project**: `pulse-expends` (type: prod)
2. **VPC**: `10.0.0.0/16` with public subnet `10.0.1.0/24`
3. **Security Group**: With all required ports (22, 80, 443, 8000, 8080, 8082, 9119)
4. **ECS Instance**: `s6.large.2` (2 vCPUs, 4GB RAM)
5. **RDS PostgreSQL**: `rds.pg.n1.large.2` with PostgreSQL 17.9
6. **OBS Buckets**: For data and document storage
7. **Elastic IP**: Static public IP for the instance
8. **KMS Key**: For encryption (if permissions allow)

### Cost Estimate (Enterprise Project):
- **ECS Instance**: ~$30/month
- **RDS PostgreSQL**: ~$50/month  
- **Elastic IP**: ~$5/month
- **OBS Storage**: ~$5/month
- **Total**: **~$90/month** (optimized for 10 users)

## 🚀 DEPLOYMENT STEPS

### Step 1: Verify IAM Permissions
Ensure IAM user has:
- `vpc:*` (full VPC permissions)
- `ecs:*` (full ECS permissions)  
- `rds:*` (full RDS permissions)
- `obs:*` (full OBS permissions)
- `kms:*` (full KMS permissions)

### Step 2: Initialize Terraform
```bash
cd /root/PulseExpends/infra
terraform init
```

### Step 3: Review Plan
```bash
terraform plan -out=tfplan-enterprise
```

### Step 4: Apply Configuration
```bash
terraform apply tfplan-enterprise
```

## 🔍 KEY DIFFERENCES FROM PREVIOUS DEPLOYMENT

| Component | Previous Deployment | Corrected Deployment |
|-----------|-------------------|-------------------|
| **Enterprise Project** | Default (ID: "0") | "pulse-expends" |
| **PostgreSQL Version** | 15 | 17.9 |
| **SSH Access** | Failed (port 22 blocked) | Will succeed (with IAM fix) |
| **RDS Creation** | Failed (IAM permissions) | Will succeed (with IAM fix) |
| **OBS Buckets** | Failed (IAM permissions) | Will succeed (with IAM fix) |
| **KMS Key** | Failed (IAM permissions) | Will succeed (with IAM fix) |

## ⚠️ PREREQUISITES BEFORE DEPLOYMENT

1. **IAM Permissions**: User must have full permissions for VPC, ECS, RDS, OBS, KMS
2. **Enterprise Project**: Must exist or user must have permission to create it
3. **Quotas**: Check service quotas in Santiago region
4. **Budget**: Ensure sufficient balance/credit

## 📊 EXPECTED OUTPUTS

After successful deployment:
- **ECS Public IP**: New Elastic IP assigned
- **RDS Endpoint**: PostgreSQL 17.9 connection string
- **SSH Access**: `ssh -i pulse-expends-key.pem ubuntu@<public-ip>`
- **Application URLs**: All services accessible on assigned IP
- **Cost Dashboard**: ~$90/month estimate

## 🔄 ROLLBACK PLAN

If deployment fails:
1. Run `terraform destroy -auto-approve`
2. Check IAM permissions
3. Verify enterprise project exists
4. Retry with corrected configuration

## ✅ READY FOR DEPLOYMENT

The configuration is now corrected to:
1. ✅ Deploy in enterprise project "pulse-expends"
2. ✅ Use PostgreSQL 17.9 (latest)
3. ✅ Include all necessary security group rules
4. ✅ Optimized for 10 concurrent users
5. ✅ Cost-effective (~$90/month)

**Next**: Run `terraform plan` to review changes, then `terraform apply` to deploy.