# Guía de Despliegue en ECS Huawei Cloud

Esta guía explica cómo desplegar PulseExpends en la instancia ECS de Huawei Cloud con configuración de subdominios.

## 📋 Prerrequisitos

1. **Instancia ECS iniciada** en Huawei Cloud Console
2. **IP Pública**: 182.160.24.205 (actual)
3. **Acceso SSH** configurado con la clave `pulse-expends-key.pem`
4. **Dominio DuckDNS**: `pulseexpends.duckdns.org` (ya configurado)

## 🚀 Pasos de Despliegue Rápido

### 1. Conectarse al servidor ECS
```bash
ssh -i pulse-expends-key.pem ubuntu@182.160.24.205
```

### 2. Clonar el repositorio
```bash
cd /opt
sudo git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends
```

### 3. Ejecutar script de configuración automática
```bash
# Hacer el script ejecutable
chmod +x setup-ecs-subdomains.sh

# Ejecutar configuración (sin token DuckDNS)
sudo ./setup-ecs-subdomains.sh

# O con token DuckDNS (recomendado)
export DUCKDNS_TOKEN="tu_token_aqui"
sudo ./setup-ecs-subdomains.sh
```

### 4. Configurar DuckDNS (si no se hizo en el paso 3)
```bash
# Editar el script con tu token
nano duckdns-config.sh

# Ejecutar configuración DNS
./duckdns-config.sh
```

## 🌐 Configuración de Subdominios

Una vez configurado, los servicios estarán disponibles en:

| Servicio | URL | Puerto | Descripción |
|----------|-----|--------|-------------|
| **Frontend Principal** | `http://pulseexpends.duckdns.org` | 80 | Aplicación web completa |
| **API MCP Server** | `http://api.pulseexpends.duckdns.org` | 8080 | API REST para transacciones |
| **PDF Parser API** | `http://pdf.pulseexpends.duckdns.org` | 8000 | Procesamiento de PDFs |
| **Status Dashboard** | `http://status.pulseexpends.duckdns.org` | 8081 | Panel de monitoreo |

## 🔧 Verificación de Servicios

### Verificar que todos los servicios estén activos:
```bash
# Ver estado de todos los servicios
sudo systemctl status nginx pulseexpends-mcp-server pulseexpends-pdf-parser pulseexpends-status

# Ver logs en tiempo real
sudo journalctl -fu nginx
sudo journalctl -fu pulseexpends-mcp-server
sudo journalctl -fu pulseexpends-pdf-parser
```

### Probar endpoints:
```bash
# Frontend
curl -I http://localhost/

# API MCP Server
curl http://localhost:8080/health

# PDF Parser
curl http://localhost:8000/health

# Status Dashboard
curl http://localhost:8081/
```

## 📁 Estructura de Directorios

```
/opt/PulseExpends/
├── frontend/                    # Código del frontend
├── backend/                     # Código del backend
├── deploy-subdomains.sh         # Script de despliegue de subdominios
├── duckdns-config.sh           # Configuración de DuckDNS
├── setup-ecs-subdomains.sh     # Configuración automática completa
└── ECS-DEPLOYMENT-GUIDE.md    # Esta guía

/var/www/
├── pulseexpends-frontend/      # Frontend desplegado
└── pulseexpends-status/        # Dashboard de status

/etc/nginx/sites-available/
└── pulseexpends               # Configuración de Nginx con subdominios

/etc/systemd/system/
├── pulseexpends-mcp-server.service
├── pulseexpends-pdf-parser.service
└── pulseexpends-status.service
```

## ⚙️ Servicios Systemd

### Comandos útiles:
```bash
# Iniciar servicios
sudo systemctl start pulseexpends-mcp-server
sudo systemctl start pulseexpends-pdf-parser
sudo systemctl start pulseexpends-status

# Habilitar inicio automático
sudo systemctl enable pulseexpends-mcp-server
sudo systemctl enable pulseexpends-pdf-parser
sudo systemctl enable pulseexpends-status

# Ver logs
sudo journalctl -u pulseexpends-mcp-server -f
sudo journalctl -u pulseexpends-pdf-parser -f

# Reiniciar servicios
sudo systemctl restart pulseexpends-mcp-server pulseexpends-pdf-parser
```

## 🔒 Seguridad y Firewall

### Puertos abiertos:
- **80/tcp** - HTTP (Nginx)
- **443/tcp** - HTTPS (pendiente de configurar SSL)
- **22/tcp** - SSH

### Configurar firewall (si es necesario):
```bash
# Permitir puertos necesarios
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable
```

## 🐛 Solución de Problemas

### 1. Nginx no inicia
```bash
# Verificar sintaxis
sudo nginx -t

# Ver logs de error
sudo tail -f /var/log/nginx/error.log

# Reiniciar Nginx
sudo systemctl restart nginx
```

### 2. Servicios Go/Python no inician
```bash
# Verificar dependencias
go version
python3 --version
pip3 list | grep -E "fastapi|uvicorn|pymupdf"

# Ver logs específicos del servicio
sudo journalctl -u pulseexpends-mcp-server --no-pager -n 50
sudo journalctl -u pulseexpends-pdf-parser --no-pager -n 50
```

### 3. Problemas con DuckDNS
```bash
# Actualizar manualmente
curl "https://www.duckdns.org/update?domains=pulseexpends&token=TU_TOKEN&ip=$(curl -s https://api.ipify.org)"

# Verificar resolución DNS
nslookup pulseexpends.duckdns.org
nslookup api.pulseexpends.duckdns.org
```

### 4. Permisos de archivos
```bash
# Corregir permisos del frontend
sudo chown -R www-data:www-data /var/www/pulseexpends-frontend
sudo chmod -R 755 /var/www/pulseexpends-frontend

# Corregir permisos del código
sudo chown -R ubuntu:ubuntu /opt/PulseExpends
```

## 📊 Monitoreo

### Dashboard de Status:
Accede a `http://status.pulseexpends.duckdns.org` para ver:
- Estado de todos los servicios
- Información del sistema
- URLs de acceso
- Comandos útiles

### Health Checks:
```bash
# Verificar salud de servicios
curl http://pulseexpends.duckdns.org/health
curl http://api.pulseexpends.duckdns.org/health
curl http://pdf.pulseexpends.duckdns.org/health
```

## 🔄 Actualizaciones

### Actualizar desde GitHub:
```bash
cd /opt/PulseExpends
sudo git pull origin main

# Reconfigurar servicios
sudo ./setup-ecs-subdomains.sh
```

### Actualizar configuración de Nginx:
```bash
sudo cp /opt/PulseExpends/backend/nginx-config/nginx-subdomains.conf /etc/nginx/sites-available/pulseexpends
sudo nginx -t
sudo systemctl reload nginx
```

## 📞 Soporte

### Logs importantes:
- Nginx: `/var/log/nginx/access.log` y `/var/log/nginx/error.log`
- Systemd: `sudo journalctl -u [servicio] -f`
- Aplicación: `/opt/PulseExpends/logs/` (si está configurado)

### Recursos:
- **Repositorio**: https://github.com/dssr1012/PulseExpends
- **Documentación**: Ver README.md en el repositorio
- **Issues**: Reportar problemas en GitHub Issues

## ✅ Verificación Final

Después del despliegue, verifica que todo funcione:

1. ✅ Frontend: http://pulseexpends.duckdns.org
2. ✅ API MCP: http://api.pulseexpends.duckdns.org/health
3. ✅ PDF Parser: http://pdf.pulseexpends.duckdns.org/health
4. ✅ Status: http://status.pulseexpends.duckdns.org
5. ✅ Servicios systemd activos
6. ✅ Nginx sirviendo correctamente
7. ✅ DNS propagado (puede tomar 5-10 minutos)

---

**Nota**: La instancia ECS debe estar en estado "RUNNING" en Huawei Cloud Console para que los servicios estén disponibles.