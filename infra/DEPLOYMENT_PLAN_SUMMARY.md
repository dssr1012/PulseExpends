# 📋 PulseExpends Infrastructure Deployment Plan
# Enterprise Project: "pulse-expends"
# Region: Santiago (la-south-2)
# PostgreSQL Version: 16 (latest available)
# =================================================

## ✅ DEPLOYMENT PLAN READY

### **Resources to Create: 25 Resources**

#### **1. ENTERPRISE PROJECT (NEW)**
- ✅ `huaweicloud_enterprise_project.pulse_expends`
  - Name: `pulse-expends`
  - Type: `prod`
  - Description: "Enterprise Project for PulseExpends application"

#### **2. NETWORKING**
- ✅ `huaweicloud_vpc.main`
  - CIDR: `10.0.0.0/16`
  - Enterprise Project: `pulse-expends`
- ✅ `huaweicloud_vpc_subnet.public`
  - CIDR: `10.0.1.0/24`
  - Gateway: `10.0.1.1`
  - DNS: `100.125.1.250`, `100.125.21.250`

#### **3. SECURITY**
- ✅ `huaweicloud_networking_secgroup.main`
  - Enterprise Project: `pulse-expends`
- ✅ **Security Group Rules (7 rules):**
  - SSH (22): Allowed from `0.0.0.0/0`
  - HTTP (80): Allowed from `0.0.0.0/0`
  - HTTPS (443): Allowed from `0.0.0.0/0`
  - MCP Server (8080): Allowed from `0.0.0.0/0`
  - Auth Server (8082): Allowed from `0.0.0.0/0`
  - PDF Parser (8000): Allowed from `0.0.0.0/0`
  - OpenClaw Dashboard (9119): Allowed from `0.0.0.0/0`
  - Egress: All outbound traffic allowed

#### **4. COMPUTE**
- ✅ `huaweicloud_compute_instance.main`
  - Type: `s6.large.2` (2 vCPUs, 4GB RAM)
  - OS: Ubuntu 22.04 server 64bit
  - Disk: 50GB system + 100GB data (SAS)
  - Key Pair: `pulse-expends-key`
  - Enterprise Project: `pulse-expends`

#### **5. DATABASE (RDS PostgreSQL 16)**
- ✅ `huaweicloud_rds_instance.postgresql`
  - Type: `rds.pg.n1.large.2` (2 vCPUs, 4GB RAM)
  - Version: **PostgreSQL 16** (latest available)
  - Storage: 50GB ESSD
  - Backup: Daily at 02:00-03:00, 7-day retention
  - Enterprise Project: `pulse-expends`
- ✅ `huaweicloud_rds_database.pulseexpends_auth`
- ✅ `huaweicloud_rds_database.pulseexpends_core`

#### **6. STORAGE**
- ✅ `huaweicloud_obs_bucket.data`
  - Bucket: `pulseexpends-data-<suffix>`
  - Storage Class: STANDARD
  - Versioning: Enabled
- ✅ `huaweicloud_obs_bucket.documents`
  - Bucket: `pulseexpends-documents-<suffix>`
  - Storage Class: STANDARD
  - Versioning: Enabled

#### **7. ENCRYPTION**
- ✅ `huaweicloud_kms_key.main`
  - Alias: `alias/pulse-expends`
  - Description: "KMS key for PulseExpends dev environment"
  - Pending Days: 7
  - Enterprise Project: `pulse-expends`

#### **8. NETWORKING SERVICES**
- ✅ `huaweicloud_vpc_eip.domain`
  - Bandwidth: 100Mbps
  - Type: 5_bgp (Dynamic BGP)
  - Purpose: Domain IP for `pulseexpends.duckdns.org`

#### **9. MONITORING**
- ✅ `huaweicloud_ces_alarm.cpu_high`
- ✅ `huaweicloud_ces_alarm.memory_high`
- ✅ `huaweicloud_ces_alarm.disk_usage_high`
- ✅ `huaweicloud_ces_alarm.rds_cpu_high`
- ✅ `huaweicloud_ces_alarm.rds_memory_high`
- ✅ `huaweicloud_ces_alarm.rds_disk_usage_high`

## 🔧 CRITICAL CHANGES FROM PREVIOUS DEPLOYMENT

### **FIXED: Enterprise Project**
- **Before**: Resources in default project (ID: "0")
- **After**: All resources in enterprise project `pulse-expends`

### **UPDATED: PostgreSQL Version**
- **Before**: PostgreSQL 15
- **After**: PostgreSQL 16 (latest available on Huawei Cloud)

### **ADDED: Security Group Rules**
- **Before**: SSH rule failed due to IAM permissions
- **After**: All ports (22, 80, 443, 8000, 8080, 8082, 9119) configured

### **ADDED: Encryption**
- KMS key for data encryption
- All resources tagged with encryption metadata

## 📊 RESOURCE SUMMARY

| Resource Type | Count | Enterprise Project | Status |
|--------------|-------|-------------------|--------|
| Enterprise Project | 1 | `pulse-expends` | ✅ NEW |
| VPC & Networking | 3 | `pulse-expends` | ✅ NEW |
| Security Group | 8 | `pulse-expends` | ✅ NEW |
| Compute (ECS) | 1 | `pulse-expends` | ✅ NEW |
| Database (RDS) | 3 | `pulse-expends` | ✅ NEW |
| Storage (OBS) | 2 | `pulse-expends` | ✅ NEW |
| Encryption (KMS) | 1 | `pulse-expends` | ✅ NEW |
| Monitoring (CES) | 6 | `pulse-expends` | ✅ NEW |
| **TOTAL** | **25** | **All in `pulse-expends`** | ✅ |

## 💰 COST ESTIMATE

| Service | Specification | Monthly Cost |
|---------|--------------|--------------|
| **ECS Instance** | s6.large.2 (2vCPU/4GB) | ~$30 |
| **RDS PostgreSQL** | rds.pg.n1.large.2 (2vCPU/4GB/50GB) | ~$50 |
| **Elastic IP** | 100Mbps bandwidth | ~$5 |
| **OBS Storage** | 2 buckets, STANDARD class | ~$5 |
| **KMS Key** | Standard encryption | ~$1 |
| **Monitoring** | 6 Cloud Eye alarms | ~$2 |
| **VPC/Network** | Basic networking | ~$2 |
| **TOTAL** | | **~$95/month** |

## 🚀 EXPECTED OUTPUTS

After deployment:
- **ECS Public IP**: New Elastic IP (will be assigned)
- **RDS Endpoint**: PostgreSQL 16 connection string
- **SSH Access**: `ssh -i pulse-expends-key.pem ubuntu@<public-ip>`
- **Application URLs**: 
  - Main: `http://pulseexpends.duckdns.org`
  - MCP Server: `http://pulseexpends.duckdns.org:8080`
  - PDF Parser: `http://pulseexpends.duckdns.org:8000`
  - Dashboard: `http://pulseexpends.duckdns.org:9119`
- **Database**: PostgreSQL 16 with 50GB storage
- **Storage**: 2 OBS buckets with versioning
- **Encryption**: KMS key for data protection
- **Monitoring**: 6 Cloud Eye alarms

## ⚠️ IAM PERMISSIONS REQUIRED

The IAM user needs these permissions:
1. `vpc:*` - Full VPC permissions
2. `ecs:*` - Full ECS permissions
3. `rds:*` - Full RDS permissions
4. `obs:*` - Full OBS permissions
5. `kms:*` - Full KMS permissions
6. `ces:*` - Full monitoring permissions

## 🔄 DEPLOYMENT COMMANDS

```bash
# Review the plan
terraform show tfplan-enterprise

# Apply the deployment
terraform apply tfplan-enterprise

# Check status
terraform output

# Destroy if needed
terraform destroy -auto-approve
```

## ✅ READY FOR DEPLOYMENT

**The infrastructure is configured to:**
1. ✅ Deploy in enterprise project `pulse-expends`
2. ✅ Use PostgreSQL 16 (latest available)
3. ✅ Include SSH access (port 22)
4. ✅ Enable encryption with KMS
5. ✅ Configure monitoring with 6 alarms
6. ✅ Optimize for 10 concurrent users
7. ✅ Estimated cost: ~$95/month

**Ready to proceed with `terraform apply`?**