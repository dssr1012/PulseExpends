# PulseExpends Infrastructure Dashboard

## 📊 Overview
A real-time dashboard for monitoring and managing the PulseExpends Huawei Cloud infrastructure.

## 🚀 Quick Start

### Option 1: Local Dashboard (Immediate)
Run the dashboard locally to view infrastructure status:

```bash
./start-dashboard.sh
```

Then open: http://localhost:8081

### Option 2: Deploy to ECS Instance
When the ECS instance is running, deploy the dashboard to the public IP:

```bash
./deploy-dashboard-to-ecs.sh
```

Then access: http://182.160.24.205

## 🖥️ Dashboard Features

### Real-time Monitoring
- **Infrastructure Status**: ECS instance, network, storage
- **Service Health**: HTTP, MCP Server, PDF Parser, SSH
- **Resource Details**: IP addresses, instance specs, security groups
- **Cost Optimization**: EIP traffic-based billing status

### Access URLs
- **Local**: http://localhost:8081
- **Network**: http://[YOUR_IP]:8081
- **ECS Instance**: http://182.160.24.205 (when deployed)
- **Domain**: http://pulseexpends.duckdns.org (when DNS configured)

### API Endpoints
- `/status` - JSON status of all infrastructure
- `/api/check/ssh` - Check SSH service (port 22)
- `/api/check/http` - Check HTTP service (port 80)
- `/api/check/mcp` - Check MCP Server (port 8080)
- `/api/check/pdf` - Check PDF Parser (port 8000)

## 🔧 Current ECS Instance Status

### ⚠️ ISSUE DETECTED
The ECS instance (`acc9edeb-1cd4-4181-8904-881135933ac5`) is currently **SHUTOFF**.

### Steps to Fix:
1. **Log into Huawei Cloud Console**
   - Go to: https://console.huaweicloud.com/ecs
   - Navigate to ECS > Instances

2. **Find the Instance**
   - Name: `pulseexpends-dev-pulse-afd05cc1-ecs`
   - ID: `acc9edeb-1cd4-4181-8904-881135933ac5`
   - IP: `182.160.24.205`

3. **Start the Instance**
   - Select the instance
   - Click "Start" button
   - Wait 2-3 minutes for boot

4. **Deploy Dashboard**
   ```bash
   ./deploy-dashboard-to-ecs.sh
   ```

## 📁 Files

### Dashboard Files
- `dashboard.html` - Main dashboard interface
- `check-status.py` - Infrastructure status checker
- `serve-dashboard.py` - HTTP server for dashboard
- `infrastructure-status.json` - Latest status data (auto-generated)

### Deployment Scripts
- `start-dashboard.sh` - Start local dashboard server
- `deploy-dashboard-to-ecs.sh` - Deploy to ECS instance
- `test-ssh.sh` - Test SSH connection to ECS

### Infrastructure Files
- `main.tf` - Terraform main configuration
- `variables.tf` - Terraform variables
- `terraform.tfvars` - Environment variables (sensitive)
- `pulse-expends-key.pem` - SSH private key

## 🛠️ Troubleshooting

### SSH Connection Issues
```bash
# Test SSH connection
./test-ssh.sh

# Debug SSH connection
ssh -i pulse-expends-key.pem -v ubuntu@182.160.24.205
```

### Service Not Responding
1. Check if instance is running in Huawei Cloud Console
2. Verify security group rules allow ports: 22, 80, 443, 8080, 8000
3. Check instance system logs in Huawei Cloud Console
4. Restart instance if needed

### Dashboard Not Accessible
1. Check if Python server is running: `ps aux | grep serve-dashboard`
2. Check firewall: `sudo ufw status`
3. Check port 8081: `netstat -tlnp | grep 8081`

## 📈 Infrastructure Details

### ECS Instance
- **Instance ID**: `acc9edeb-1cd4-4181-8904-881135933ac5`
- **Type**: `ac8.large.2` (2 vCPU, 8GB RAM)
- **OS**: Ubuntu 22.04
- **Public IP**: `182.160.24.205`
- **Private IP**: `10.0.1.108`

### Network
- **VPC**: `10.0.0.0/16`
- **Subnet**: `10.0.1.0/24`
- **Security Group**: Allows 22, 80, 443, 8080, 8000
- **EIP**: `182.160.24.205` (Traffic billing mode)

### Storage
- **OBS Data Bucket**: `pulseexpends-data-dev-pulse-afd05cc1`
- **OBS Documents Bucket**: `pulseexpends-documents-dev-pulse-afd05cc1`
- **KMS Key**: `2a27ce80-f6a2-47a6-a191-47f8ac07ccd0`

## 🔄 Updates & Maintenance

### Update Dashboard
1. Edit `dashboard.html`, `check-status.py`, or `serve-dashboard.py`
2. Test locally: `./start-dashboard.sh`
3. Deploy to ECS: `./deploy-dashboard-to-ecs.sh`

### Update Infrastructure
1. Modify Terraform files (`main.tf`, `variables.tf`)
2. Plan changes: `./terraform plan`
3. Apply changes: `./terraform apply`

### Check Status
```bash
# Run status checker
python3 check-status.py

# View JSON status
cat infrastructure-status.json
```

## 📞 Support

### Immediate Issues
1. Check Huawei Cloud Console for instance status
2. Review security group rules
3. Check system logs

### Documentation
- Terraform State: `./terraform show`
- Outputs: `./terraform output`
- GitHub Repo: https://github.com/dssr1012/PulseExpends-Infra

### Cost Optimization
✅ **EIP Billing Mode Changed**: From "bandwidth" to "traffic"
- **Before**: Fixed cost for 300 Mbps bandwidth
- **After**: Pay-per-use based on actual traffic
- **Savings**: Significant reduction for variable workloads

---

**Last Updated**: 2026-05-19  
**Terraform State**: Applied  
**ECS Status**: SHUTOFF (needs manual start)  
**Dashboard**: Ready for deployment