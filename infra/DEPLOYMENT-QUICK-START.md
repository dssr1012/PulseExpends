# 🚀 Despliegue Rápido - PulseExpends

Guía paso a paso para desplegar la aplicación completa una vez que la infraestructura esté lista.

## 📋 Prerrequisitos

1. **Infraestructura Huawei Cloud** desplegada y funcionando
2. **IP Pública** asignada a la instancia ECS
3. **Acceso SSH** configurado con la clave `pulse-expends-key.pem`
4. **Dominio DuckDNS** configurado para apuntar a la IP pública

## 🔧 Pasos de Despliegue

### 1. Obtener IP Pública y Credenciales SSH
```bash
# Desde el directorio de Terraform
cd /root/PulseExpends-Infra

# Obtener la IP pública
terraform output domain_ip_address

# Obtener comando SSH
terraform output ssh_connection_command
```

### 2. Conectarse al Servidor ECS
```bash
# Usar el comando SSH del output o:
ssh -i pulse-expends-key.pem ubuntu@<IP_PUBLICA>
```

### 3. Clonar el Repositorio en el Servidor
```bash
# En el servidor ECS:
cd /opt
sudo git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends
```

### 4. Configurar DuckDNS (si no está configurado)
```bash
# En el servidor ECS:
cd /opt/PulseExpends/infra/scripts

# Editar el script con tu token DuckDNS
nano configure-duckdns.sh

# Ejecutar configuración
./configure-duckdns.sh
```

### 5. Desplegar la Aplicación con Paths
```bash
# En el servidor ECS:
cd /opt/PulseExpends

# Hacer ejecutable el script
chmod +x infra/deploy-with-paths.sh

# Ejecutar despliegue completo
sudo ./infra/deploy-with-paths.sh
```

### 6. Verificar el Despliegue
```bash
# En el servidor ECS:
cd /opt/PulseExpends

# Ejecutar script de verificación
./infra/test-paths-config.sh

# Verificar servicios
sudo systemctl status nginx pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service
```

## 🌐 URLs de Acceso

Una vez desplegado, accede a:

| Servicio | URL | Descripción |
|----------|-----|-------------|
| **Frontend Principal** | `http://pulseexpends.duckdns.org/` | Aplicación web completa |
| **API MCP Server** | `http://pulseexpends.duckdns.org/mcp/` | API REST para transacciones |
| **PDF Parser API** | `http://pulseexpends.duckdns.org/pdf/` | Procesamiento de PDFs |
| **Status Dashboard** | `http://pulseexpends.duckdns.org/status/` | Panel de monitoreo |

## 🧪 Health Checks

Verifica que todo funcione:

```bash
# Health checks locales
curl http://localhost/health
curl http://localhost:8080/health
curl http://localhost:8000/health
curl http://localhost:8081/

# Health checks públicos (después de configurar DNS)
curl http://pulseexpends.duckdns.org/health
curl http://pulseexpends.duckdns.org/mcp/health
curl http://pulseexpends.duckdns.org/pdf/health
curl http://pulseexpends.duckdns.org/status/health
```

## ⚙️ Servicios Systemd

### Comandos útiles:
```bash
# Ver estado de todos los servicios
sudo systemctl status nginx pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service

# Iniciar servicios
sudo systemctl start pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service

# Habilitar inicio automático
sudo systemctl enable pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service

# Ver logs
sudo journalctl -u pulseexpends-mcp.service -f
sudo journalctl -u pulseexpends-pdf-parser.service -f
sudo journalctl -u pulseexpends-status.service -f
sudo journalctl -u nginx -f

# Reiniciar todos los servicios
sudo systemctl restart nginx pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service
```

## 🐛 Solución de Problemas Comunes

### 1. Nginx no inicia
```bash
# Verificar sintaxis
sudo nginx -t

# Ver logs de error
sudo tail -f /var/log/nginx/error.log

# Reiniciar Nginx
sudo systemctl restart nginx
```

### 2. Servicios no responden
```bash
# Verificar que los servicios estén corriendo
sudo systemctl status pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service

# Verificar puertos
sudo netstat -tlnp | grep -E "8080|8000|8081"

# Ver logs específicos
sudo journalctl -u pulseexpends-mcp.service --no-pager -n 50
sudo journalctl -u pulseexpends-pdf-parser.service --no-pager -n 50
```

### 3. Problemas con DuckDNS
```bash
# Actualizar manualmente
curl "https://www.duckdns.org/update?domains=pulseexpends&token=TU_TOKEN&ip=$(curl -s https://api.ipify.org)"

# Verificar resolución DNS
nslookup pulseexpends.duckdns.org
```

### 4. Permisos de archivos
```bash
# Corregir permisos del frontend
sudo chown -R www-data:www-data /var/www/pulseexpends-frontend
sudo chmod -R 755 /var/www/pulseexpends-frontend

# Corregir permisos del código
sudo chown -R ubuntu:ubuntu /opt/PulseExpends
```

## 🔄 Actualizaciones

### Actualizar desde GitHub:
```bash
cd /opt/PulseExpends
sudo git pull origin main

# Reconfigurar servicios
sudo ./infra/deploy-with-paths.sh
```

### Actualizar solo configuración Nginx:
```bash
sudo cp /opt/PulseExpends/infra/src/nginx/pulseexpends.conf /etc/nginx/sites-available/pulseexpends
sudo nginx -t
sudo systemctl reload nginx
```

## 📊 Monitoreo

### Dashboard de Status:
Accede a `http://pulseexpends.duckdns.org/status/` para:
- Ver estado de todos los servicios
- Información del sistema
- URLs de acceso
- Comandos útiles

### Logs importantes:
- Nginx: `/var/log/nginx/access.log` y `/var/log/nginx/error.log`
- Systemd: `sudo journalctl -u [servicio] -f`
- Aplicación: logs específicos de cada servicio

## ✅ Checklist de Verificación Final

- [ ] Instancia ECS en estado "RUNNING"
- [ ] DNS DuckDNS configurado y propagado
- [ ] Script `deploy-with-paths.sh` ejecutado exitosamente
- [ ] Todos los servicios systemd activos
- [ ] Nginx sirviendo en puerto 80
- [ ] Frontend accesible: `curl -I http://pulseexpends.duckdns.org/`
- [ ] APIs responden: `curl http://pulseexpends.duckdns.org/mcp/health`
- [ ] Status Dashboard funciona: `curl http://pulseexpends.duckdns.org/status/`

## 📞 Soporte

### Recursos:
- **Repositorio**: https://github.com/dssr1012/PulseExpends
- **Documentación**: Ver `README.md` en el repositorio
- **Issues**: Reportar problemas en GitHub Issues

### Comandos de diagnóstico:
```bash
# Verificar estado completo
cd /opt/PulseExpends
./infra/test-paths-config.sh

# Verificar conectividad
ping -c 4 pulseexpends.duckdns.org

# Verificar puertos abiertos
sudo netstat -tlnp

# Verificar uso de recursos
top -b -n 1 | head -20
free -h
df -h
```

---

**Nota**: La propagación DNS puede tomar 5-10 minutos después de configurar DuckDNS. Si los servicios no son accesibles inmediatamente, espera unos minutos y vuelve a intentar.