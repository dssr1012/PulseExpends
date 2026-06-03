# Guía de Despliegue con Paths (Versión Estable)

Esta guía explica cómo desplegar PulseExpends usando paths en lugar de subdominios. Esta es la **versión estable** que functiona con DuckDNS y cualquier service DNS.

## 🎯 ¿Por qué Paths en lugar de Subdominios?

### **Problema con DuckDNS:**
- DuckDNS no soporta subdominios de forma nativa
- Requeriría create dominios separados (`api-pulseexpends.duckdns.org`, etc.)
- Más complejo de configurar y mantener

### **Solución con Paths:**
- ✅ Funciona con cualquier DNS (incluido DuckDNS)
- ✅ Más simple de configurar
- ✅ Fácil de migrar a Cloudflare después
- ✅ Menos dependencias

## 🌐 URLs de Acceso

| Path | service | Puerto Interno | Descripción |
|------|----------|----------------|-------------|
| **`http://pulseexpends.duckdns.org/`** | Frontend Principal | 80 | Aplicación web completa |
| **`http://pulseexpends.duckdns.org/mcp/`** | MCP Server api | 8080 | api REST para transactiones |
| **`http://pulseexpends.duckdns.org/pdf/`** | PDF Parser api | 8000 | Procesamiento de PDFs |
| **`http://pulseexpends.duckdns.org/status/`** | Status Dashboard | 8081 | Panel de monitoreo |

## 📋 Prerrequisitos

1. **Instancia ECS iniciada** en Huawei Cloud Console
2. **IP Pública**: 182.160.24.205 (actual)
3. **Acceso SSH** configurado con la clave `pulse-expends-key.pem`
4. **Dominio DuckDNS**: `pulseexpends.duckdns.org` (ya configurado)

## 🚀 Despliegue Rápido

### 1. Conectarse al servidor ECS
```bash
ssh -i pulse-expends-key.pem ubuntu@182.160.24.205
```

### 2. Clonar el repository (si no está clonado)
```bash
cd /opt
sudo git clone https://github.com/dssr1012/PulseExpends.git
cd PulseExpends
```

### 3. Ejecutar script de configuration automática
```bash
# Hacer el script ejecutable
chmod +x infra/deploy-with-paths.sh

# Ejecutar configuration completa
sudo ./infra/deploy-with-paths.sh
```

### 4. Configurar DuckDNS (si no está configurado)
```bash
# Solo necesitas configurar el dominio principal
# Ya debería estar configurado para apuntar a 182.160.24.205

# Verificar configuration DNS
nslookup pulseexpends.duckdns.org
# Debería devolver: 182.160.24.205
```

## 🔧 configuration Manual

### Estructura de Directorios:
```
/opt/PulseExpends/
├── infra/
│   ├── deploy-with-paths.sh          # Script de despliegue
│   ├── src/nginx/pulseexpends.conf   # configuration Nginx
│   └── src/scripts/                  # Servicios systemd
├── services/
│   ├── backend/                      # Backend (MCP Server, PDF Parser)
│   └── frontend/                     # Frontend web
└── infra/dashboard/                  # Status Dashboard
```

### configuration Nginx (`/etc/nginx/sites-available/pulseexpends`):
```nginx
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;
    
    # Frontend React App
    location / {
        root /var/www/pulseexpends-frontend;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # MCP Server api
    location /mcp/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
    }
    
    # PDF Parser api
    location /pdf/ {
        proxy_pass http://localhost:8000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
    }
    
    # Status Dashboard api
    location /status/ {
        proxy_pass http://localhost:8081/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
    }
    
    # Health checks
    location /health {
        return 200 'OK';
        add_header Content-Type text/plain;
    }
}
```

## ⚙️ Servicios Systemd

### Servicios configurados:
1. **`pulseexpends-mcp.service`** - MCP Server (puerto 8080)
2. **`pulseexpends-pdf-parser.service`** - PDF Parser (puerto 8000)
3. **`pulseexpends-status.service`** - Status Dashboard (puerto 8081)
4. **`nginx`** - Servidor web (puerto 80)

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

## 🧪 Verificación

### 1. Verificar servicios locales:
```bash
# Frontend (debe servir index.html)
curl -I http://localhost/

# MCP Server
curl http://localhost:8080/health

# PDF Parser
curl http://localhost:8000/health

# Status Dashboard
curl http://localhost:8081/
```

### 2. Verificar configuration Nginx:
```bash
# Verificar sintaxis
sudo nginx -t

# Verificar configuration cargada
sudo nginx -T | grep -A 20 "server_name _"

# Ver logs de acceso
sudo tail -f /var/log/nginx/access.log

# Ver logs de error
sudo tail -f /var/log/nginx/error.log
```

### 3. Verificar desde fuera:
```bash
# Reemplaza con tu IP pública o dominio
DOMAIN="pulseexpends.duckdns.org"

# Frontend
curl -I http://$DOMAIN/

# APIs
curl http://$DOMAIN/mcp/health
curl http://$DOMAIN/pdf/health
curl http://$DOMAIN/status/

# Health check general
curl http://$DOMAIN/health
```

## 🐛 Solución de Problemas

### Problema 1: Nginx no inicia
```bash
# Verificar sintaxis
sudo nginx -t

# Ver logs de error
sudo journalctl -u nginx --no-pager -n 50

# Reiniciar Nginx
sudo systemctl restart nginx
```

### Problema 2: Servicios no responden
```bash
# Verificar que los servicios estén corriendo
sudo systemctl status pulseexpends-mcp.service pulseexpends-pdf-parser.service pulseexpends-status.service

# Verificar puertos
sudo netstat -tlnp | grep -E "8080|8000|8081"

# Verificar logs de cada service
sudo journalctl -u pulseexpends-mcp.service --no-pager -n 50
sudo journalctl -u pulseexpends-pdf-parser.service --no-pager -n 50
sudo journalctl -u pulseexpends-status.service --no-pager -n 50
```

### Problema 3: Permisos de archivos
```bash
# Corregir permisos del frontend
sudo chown -R www-data:www-data /var/www/pulseexpends-frontend
sudo chmod -R 755 /var/www/pulseexpends-frontend

# Corregir permisos del código
sudo chown -R ubuntu:ubuntu /opt/PulseExpends
```

### Problema 4: DNS no resuelve
```bash
# Verificar resolución DNS
nslookup pulseexpends.duckdns.org

# update DuckDNS manualmente
curl "https://www.duckdns.org/update?domains=pulseexpends&token=TU_TOKEN&ip=$(curl -s https://api.ipify.org)"
```

## 🔄 Actualizaciones

### update desde GitHub:
```bash
cd /opt/PulseExpends
sudo git pull origin main

# Reconfigurar servicios
sudo ./infra/deploy-with-paths.sh
```

### update solo configuration Nginx:
```bash
sudo cp /opt/PulseExpends/infra/src/nginx/pulseexpends.conf /etc/nginx/sites-available/pulseexpends
sudo nginx -t
sudo systemctl reload nginx
```

## 📊 Monitoreo

### Dashboard de Status:
Accede a `http://pulseexpends.duckdns.org/status/` para ver:
- Estado de todos los servicios
- information del sistema
- URLs de acceso
- Comandos útiles

### Health Checks:
```bash
# Verificar salud de servicios
curl http://pulseexpends.duckdns.org/health
curl http://pulseexpends.duckdns.org/mcp/health
curl http://pulseexpends.duckdns.org/pdf/health
curl http://pulseexpends.duckdns.org/status/health
```

## 🔒 Seguridad

### Puertos abiertos:
- **80/tcp** - HTTP (Nginx)
- **22/tcp** - SSH (acceso remoto)

### Recomendaciones de seguridad:
1. **Configurar firewall** (si no está configurado):
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 22/tcp
   sudo ufw enable
   ```

2. **Configurar SSL/HTTPS** (para producción):
   - Usar Let's Encrypt con Certbot
   - update Nginx para usar puerto 443

3. **Restringir acceso SSH**:
   - Usar claves SSH en lugar de passwords
   - Cambiar puerto SSH por defecto
   - Usar fail2ban

## 🚀 Migración a Cloudflare (Futuro)

Cuando estés listo para migrar a Cloudflare:

1. **Registrar dominio** en Cloudflare (o transferir uno existente)
2. **Configurar DNS** en Cloudflare:
   ```
   pulseexpends.tudominio.com A → 182.160.24.205
   ```
3. **update configuration Nginx** para usar el nuevo dominio
4. **Configurar SSL** con Cloudflare (gratis)
5. **Habilitar CDN** y otras características de Cloudflare

### Ventajas de Cloudflare:
- ✅ Subdominios ilimitados
- ✅ SSL/TLS gratuito
- ✅ CDN global
- ✅ Protección DDoS
- ✅ Analytics y caching

## ✅ Checklist de Verificación

- [ ] Instancia ECS en estado "RUNNING"
- [ ] DNS resuelve correctamente: `nslookup pulseexpends.duckdns.org`
- [ ] Nginx sirviendo en puerto 80: `sudo netstat -tlnp | grep :80`
- [ ] Servicios systemd activos: `sudo systemctl status nginx pulseexpends-*`
- [ ] Frontend accesible: `curl -I http://pulseexpends.duckdns.org/`
- [ ] APIs responden: `curl http://pulseexpends.duckdns.org/mcp/health`
- [ ] Status Dashboard functiona: `curl http://pulseexpends.duckdns.org/status/`

## 📞 Soporte

### Logs importantes:
- Nginx: `/var/log/nginx/access.log` y `/var/log/nginx/error.log`
- Systemd: `sudo journalctl -u [service] -f`
- Aplicación: logs específicos de cada service

### Recursos:
- **repository**: https://github.com/dssr1012/PulseExpends
- **Documentación**: Ver `README.md` en el repository
- **Issues**: Reportar problemas en GitHub Issues

---

**Nota**: La instancia ECS debe estar en estado "RUNNING" en Huawei Cloud Console para que los servicios estén disponibles. Si la instancia está apagada, iníciala desde la consola primero.