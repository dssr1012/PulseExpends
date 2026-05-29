# 🚀 PulseExpends Deployment Checklist - Santiago Region

## ✅ **Phase 1G - Frontend Development: 100% COMPLETE**
- 7 React components built with TypeScript
- 6 Zustand stores for state management
- 109 API endpoints fully integrated
- Complete UI/UX with Tailwind CSS

## 🔧 **Infrastructure Status: CLEAN**
- Terraform resources: **0** (all destroyed)
- Huawei Cloud account: Restricted (cannot create new resources)
- ECS instance: Needs to be started in Santiago region

## 🔑 **SSH Key Generated:**
**Private Key:** `/root/PulseExpends/infra/pulse-expends-key.pem`
**Public Key:** Generated and ready for import

## 📋 **Immediate Actions Required:**

### **1. Huawei Cloud Console - Start ECS Instance**
1. **Log into** Huawei Cloud Console
2. **Navigate to** ECS → Instances
3. **Filter by region:** `la-south-2` (Santiago)
4. **Find the instance** (look for "pulseexpends" in name)
5. **Start the instance** if it's stopped
6. **Note the public IP address**

### **2. Import SSH Key to Huawei Cloud**
1. **Go to** ECS → Key Pairs
2. **Click** "Import Key Pair"
3. **Name:** `pulse-expends-key`
4. **Paste Public Key:**
```
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDHp+gf4OH39uJONLLelx8ldn3BtrdnJJAKpXu9QkdK7dBRa/gezHziwZaPrXmC/GK7TTraXMNbERPXVIYXVSdJxMAydp+raHy0f0WrOTJHvdc5PPxBsqaEGFn3U4OrloC0DNv7asUg+sorZ/q6sCs1tN7nFuIsr7gjkG8KHy7Jai44ARUwXk7fFw9A0R6mMpQ2cbNUEsSUP1Z4tqALmnq4Gt7irMCOJRIkHbMf5AzePYuUCRkD251xlTHZ3n1xgQZSyygeJwjBtw8P8p50hDBF12ofjrAP6DxqGj6VR7iNar2E0fToh//Tb535IbBef9FhFOg6eC62KbyMK8Tcdx+2kicRa2oj4iXFApVx32ROOcUENPOwoJkwhWc44OtjJhhCL7f/Z/NkfmiTlLhtbK8MRryHQrsBd6LZoLWDEPMUdOKPKJdd+7dagG658wt5GmkVd7o6pSyNLBPfwSxKb0YVVGL2FxaT0zmUBjEyTSsx6ECt75wRIp+6xaIxnGR+1dRi5rVwSE1XhHx/cR0rAXYTsgPsuDU2JR6AjUxrgmtw6/Dq+zl4KzIViLFXwXlSDbGT9D2cv7QRBFHrCMY+fGj5iIJbjWfRBWYXC9q889rcfNErTcOKlBVXGj+Tp/VZsSp8B+J56JK5u3tZQRHi70zfyTdeck+wO5Rujmk/pnSUrQ== root@clawdbot-demo
```
5. **Click** "OK"

### **3. Bind SSH Key to ECS Instance**
1. **Go back to** ECS → Instances
2. **Select your instance**
3. **Click** "More" → "Change Key Pair"
4. **Select** `pulse-expends-key`
5. **Click** "OK" (instance will reboot)

### **4. Update Deployment Script**
```bash
cd /root/PulseExpends/infra
# Replace IP with your instance's public IP
sed -i 's/IP="182.160.24.205"/IP="YOUR_NEW_IP_ADDRESS"/' deploy-app.sh
```

### **5. Deploy Application**
```bash
cd /root/PulseExpends/infra
chmod +x deploy-app.sh
./deploy-app.sh
```

## 🎯 **Services to be Deployed:**

### **Backend Services:**
- **Auth API** (port 8082) - JWT authentication, Google OAuth
- **Main API** (port 8083) - 6 services, 109 endpoints
- **MCP Server** (port 8080) - Existing service
- **PDF Parser** (port 8000) - Existing service

### **Frontend Services:**
- **React App** (port 80/443) - 7 components, 6 Zustand stores
- **Nginx Reverse Proxy** - Routes traffic to services

### **Database:**
- **PostgreSQL RDS** - Existing database in Santiago region

## 📊 **Deployment Verification:**

After running `./deploy-app.sh`, check these endpoints:
1. **Frontend:** `http://YOUR_IP`
2. **Auth API:** `http://YOUR_IP:8082/health`
3. **Main API:** `http://YOUR_IP:8083/health`
4. **MCP Server:** `http://YOUR_IP:8080/health`
5. **PDF Parser:** `http://YOUR_IP:8000/health`

## ⚠️ **Troubleshooting:**

### **If SSH connection fails:**
1. **Check instance status** - must be "ACTIVE"
2. **Verify security group** allows port 22 (SSH)
3. **Confirm key binding** - key must be bound to instance
4. **Wait for reboot** - instance reboots after key change

### **If deployment script fails:**
1. **Check disk space** on ECS instance
2. **Verify Docker** is installed
3. **Check network connectivity** to Docker Hub
4. **Review logs:** `docker-compose logs -f`

## 🎉 **Deployment Complete When:**
- All 6 services are running (`docker-compose ps`)
- Frontend is accessible on port 80
- All API endpoints respond with 200 OK
- Database connections are established

## 📞 **Support:**
If issues persist, check:
- `/root/PulseExpends/infra/deploy-app.sh` logs
- Docker container logs
- Huawei Cloud console for instance metrics
- Security group rules in VPC

**Estimated deployment time:** 20-25 minutes after instance is running