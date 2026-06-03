# 🚀 Guía Rápida para Reactivar la Infraestructura

## ⚡ Comandos Rápidos

### Opción 1: Script Automático (Recomendado)
```bash
cd /root/PulseExpends-Infra
./reactivate-infrastructure.sh
```

### Opción 2: Pasos Manuales
```bash
# 1. Ir al directorio de infraestructura
cd /root/PulseExpends-Infra

# 2. Verificar credenciales
cp terraform.tfvars.example terraform.tfvars
# EDITAR terraform.tfvars con tus credenciales reales

# 3. Inicializar y aplicar
terraform init
terraform apply -auto-approve

# 4. get IP pública
terraform output application_url

# 5. Conectar al servidor
ssh -i pulse-expends-key.pem root@<IP-PÚBLICA>

# 6. Desplegar application
cd /opt/PulseExpends
./setup-ecs-subdomains.sh
```

## 📋 Qué se Reactivará

### Recursos Principales
1. **ECS Instance**: `ac8.large.2` (2 vCPU, 8GB RAM, Ubuntu 20.04)
2. **EIP (IP Pública)**: Nueva IP elástica
3. **VPC**: Red `10.0.0.0/16`
4. **Security Group**: Reglas para puertos 22, 80, 443, 8080, 8000, 8082, 18789
5. **OBS Buckets**: 2 buckets para datos y documentos
6. **KMS Key**: Clave de encriptación

### Tiempo Estimado
- **Terraform apply**: 3-5 minutos
- **Inicio del servidor**: 2-3 minutos
- **Despliegue application**: 5-10 minutos
- **Total**: 10-18 minutos

## 🔧 configuration Post-Reactivar

### 1. Configurar DuckDNS
```bash
# En el servidor ECS
cd /opt/PulseExpends
nano duckdns-update.sh
# update token DuckDNS
./configure-duckdns-secure.sh
```

### 2. Verificar Servicios
```bash
# En el servidor ECS
systemctl status pulseexpends-auth.service
systemctl status pulseexpends-mcp.service
systemctl status pulseexpends-pdf.service
systemctl status nginx
```

### 3. Probar Acceso
```bash
# Desde tu máquina local
curl http://<IP-PÚBLICA>/
curl http://<IP-PÚBLICA>/api/auth/health
curl http://<IP-PÚBLICA>/api/mcp/health
curl http://<IP-PÚBLICA>/api/pdf/health
```

## 🗄️ RDS PostgreSQL (Opcional)

### Para agregar base de datos gestionada:
```bash
# 1. create configuration RDS
cd /root/PulseExpends-Infra
cp rds-example.tfvars rds.auto.tfvars
# EDITAR con credenciales seguras

# 2. Aplicar RDS
terraform apply -var-file="rds.auto.tfvars"

# 3. Configurar application
# En el servidor ECS:
echo "DATABASE_URL=postgresql://user:password@rds-endpoint:5432/pulseexpends" > /opt/PulseExpends/backend/auth/.env
sudo systemctl restart pulseexpends-auth.service
```

## 🚨 Solución de Problemas Comunes

### Problema: No se puede conectar por SSH
```bash
# Verificar que la IP sea correcta
terraform output eip_address

# Verificar grupo de seguridad
# Debe tener regla para puerto 22 desde 0.0.0.0/0

# Verificar clave SSH
ls -la pulse-expends-key.pem
chmod 600 pulse-expends-key.pem
```

### Problema: Servicios no inician
```bash
# En el servidor ECS
journalctl -u pulseexpends-auth.service --no-pager -l
journalctl -u nginx --no-pager -l

# Reiniciar servicios
sudo systemctl restart pulseexpends-auth.service pulseexpends-mcp.service pulseexpends-pdf.service nginx
```

### Problema: DuckDNS no actualiza
```bash
# Verificar script
cat /opt/PulseExpends/duckdns-update.sh

# Ejecutar manualmente
/opt/PulseExpends/duckdns-update.sh

# Verificar logs
tail -f /var/log/duckdns.log
```

## 📊 Monitoreo Post-Reactivar

### Comandos Útiles
```bash
# Ver uso de recursos
top
htop
df -h
free -h

# Ver logs en tiempo real
journalctl -f -u pulseexpends-auth.service
journalctl -f -u nginx

# Ver acceso web
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

### Verificación Completa
```bash
# Script de verification
cd /opt/PulseExpends
./check-status.sh
```

## 💾 Backup del Estado

### Archivos Importantes
```
/root/PulseExpends-Infra/
├── terraform.tfstate.backup.20260520_090453  # Backup del estado anterior
├── terraform.tfstate                          # Estado actual (vacío)
├── reactivate-infrastructure.sh              # Script de reactivación
└── INFRASTRUCTURE-SHUTDOWN-SUMMARY.md        # Resumen del apagado
```

### Restaurar desde Backup (si es necesario)
```bash
cp terraform.tfstate.backup.20260520_090453 terraform.tfstate
terraform init
terraform plan
```

## 📞 Soporte

### Documentación
- `RDS-README.md` - Guía completa de RDS PostgreSQL
- `README.md` - Documentación del proyecto
- `DASHBOARD-README.md` - Guía del dashboard
- `ECS-DEPLOYMENT-GUIDE.md` - Guía de despliegue

### Repositorios
- **Infraestructura**: https://github.com/dssr1012/PulseExpends-Infra
- **Aplicación**: https://github.com/dssr1012/PulseExpends

### Archivos de configuration
- `main.tf` - configuration principal de Terraform
- `variables.tf` - Variables de configuration
- `outputs.tf` - Outputs de Terraform
- `rds.tf` - Módulo RDS PostgreSQL

---

**¡Listo para reactivar cuando sea necesario!** 🚀

**Tiempo estimado para tener TODO functionando:** 15-20 minutos
**Costo estimado mensual:** $85-130 (dependiendo del tráfico)
**Estado actual:** ✅ configuration preservada, lista para reactivar