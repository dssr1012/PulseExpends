# 📋 DIAGNÓSTICO Y SOLUCIÓN - PulseExpends

## 🚨 PROBLEMA IDENTIFICADO

**Nginx está mal configurado en el servidor ECS (`182.160.24.205`)**.

### Síntomas:
1. ✅ **Servidor ECS activo**: Responde en puertos 80, 8080, 8000
2. ✅ **Frontend funcionando**: http://182.160.24.205/ muestra la aplicación
3. ✅ **APIs funcionando**:
   - MCP Server API (puerto 8080): ✅ Saludable (`{"status":"healthy"}`)
   - PDF Parser API (puerto 8000): ✅ Saludable (`{"status":"healthy"}`)
4. ❌ **Path-based routing NO funciona**: 
   - `http://182.160.24.205/mcp/health` → Devuelve el frontend (debería redirigir a puerto 8080)
   - `http://182.160.24.205/pdf/health` → Devuelve el frontend (debería redirigir a puerto 8000)
   - `http://182.160.24.205/status/` → Devuelve el frontend (puerto 8081 no responde)
5. ❌ **Status Dashboard no funciona**: Puerto 8081 no responde

### Root Cause:
La configuración de nginx en el servidor no tiene las reglas de `location` para redirigir `/mcp/*`, `/pdf/*` y `/status/*` a los servicios correspondientes. En su lugar, está sirviendo el frontend en **TODAS** las rutas.

## 🔍 EVIDENCIA

### Servicios activos:
```bash
# APIs funcionando en puertos directos
curl http://182.160.24.205:8080/health      # ✅ {"status":"healthy"}
curl http://182.160.24.205:8000/health      # ✅ {"status":"healthy"}

# Pero las rutas nginx no redirigen
curl http://182.160.24.205/mcp/health       # ❌ Devuelve HTML del frontend
curl http://182.160.24.205/pdf/health       # ❌ Devuelve HTML del frontend
```

### Configuración correcta necesaria (en `/etc/nginx/sites-available/pulseexpends`):
```nginx
server {
    listen 80 default_server;
    server_name _;
    
    # Frontend en raíz
    location / {
        root /var/www/pulseexpends-frontend;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # API MCP Server
    location /mcp/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # API PDF Parser
    location /pdf/ {
        proxy_pass http://localhost:8000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # Status Dashboard
    location /status/ {
        proxy_pass http://localhost:8081/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🛠️ SOLUCIONES

### SOLUCIÓN 1: Corregir nginx en el servidor (Recomendada)

**Requisito**: Acceso SSH al servidor ECS (actualmente bloqueado por problemas de clave)

**Pasos**:
```bash
# 1. Conectarse al servidor (si se resuelve el problema de SSH)
ssh -i pulse-expends-key.pem ubuntu@182.160.24.205

# 2. Ejecutar script de despliegue
cd /opt/PulseExpends
sudo ./infra/deploy-with-paths.sh

# O manualmente:
sudo cp /opt/PulseExpends/infra/src/nginx/pulseexpends.conf /etc/nginx/sites-available/pulseexpends
sudo ln -sf /etc/nginx/sites-available/pulseexpends /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

### SOLUCIÓN 2: Workaround - Frontend modificado (Implementada)

**Ya implementada**: Modifiqué el frontend para usar puertos directos:

```javascript
// ANTES (no funciona por nginx mal configurado):
const API_BASE_URL = 'http://182.160.24.205/mcp';
const PDF_API_URL = 'http://182.160.24.205/pdf';

// DESPUÉS (funciona con puertos directos):
const API_BASE_URL = 'http://182.160.24.205:8080';
const PDF_API_URL = 'http://182.160.24.205:8000';
```

**Archivo modificado**: `/root/PulseExpends/infra/frontend/index.html`

### SOLUCIÓN 3: Proxy inverso local (Para desarrollo/pruebas)

**Script creado**: `/root/PulseExpends/test-proxy.py`

**Ejecutar**:
```bash
cd /root/PulseExpends
python3 test-proxy.py
```

**Acceder**: http://localhost:8888/

**Routing**:
- `/` → Frontend (puerto 80)
- `/mcp/*` → MCP Server API (puerto 8080)
- `/pdf/*` → PDF Parser API (puerto 8000)
- `/status/*` → Status Dashboard (puerto 8081, actualmente no funciona)

### SOLUCIÓN 4: Configurar nginx local (Para testing completo)

**Configuración**: `/root/PulseExpends/nginx-workaround.conf`

**Instalar y configurar**:
```bash
# Instalar nginx
sudo apt-get update
sudo apt-get install -y nginx

# Copiar configuración
sudo cp /root/PulseExpends/nginx-workaround.conf /etc/nginx/sites-available/pulseexpends-workaround
sudo ln -s /etc/nginx/sites-available/pulseexpends-workaround /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Probar y recargar
sudo nginx -t
sudo systemctl reload nginx
```

**Acceder**: http://localhost:8088/

## 📊 ESTADO ACTUAL

### ✅ Funcionando:
1. **Frontend**: http://182.160.24.205/
2. **MCP Server API**: http://182.160.24.205:8080/
3. **PDF Parser API**: http://182.160.24.205:8000/
4. **Frontend modificado**: Usa puertos directos (8080, 8000)

### ❌ No funcionando:
1. **Path-based routing**: `/mcp/*`, `/pdf/*`, `/status/*`
2. **Status Dashboard**: Puerto 8081 no responde
3. **Acceso SSH**: Problemas con la clave

### 🔧 Soluciones implementadas:
1. **Frontend modificado** para usar puertos directos ✓
2. **Script de proxy local** para testing ✓
3. **Configuración nginx local** lista para usar ✓

## 🚀 PASOS SIGUIENTES

### Inmediato (Workaround):
1. **Usar frontend modificado** con puertos directos
2. **Ejecutar proxy local** para testing completo
3. **Verificar que todas las funcionalidades funcionan**

### A corto plazo (Corrección en servidor):
1. **Resolver problema de SSH** con la clave `pulse-expends-key.pem`
2. **Ejecutar script de despliegue** en el servidor: `./infra/deploy-with-paths.sh`
3. **Verificar configuración nginx** con `sudo nginx -t`
4. **Recargar nginx**: `sudo systemctl reload nginx`

### A mediano plazo:
1. **Revisar Status Dashboard** (puerto 8081)
2. **Configurar SSL/HTTPS**
3. **Configurar dominio DuckDNS** correctamente
4. **Implementar monitoreo** y alertas

## 📁 ARCHIVOS CREADOS

1. **`/root/PulseExpends/infra/frontend/index.html`** - Frontend modificado (usa puertos directos)
2. **`/root/PulseExpends/test-proxy.py`** - Proxy inverso local para testing
3. **`/root/PulseExpends/nginx-workaround.conf`** - Configuración nginx local
4. **`/root/PulseExpends/test-api-direct-ports.html`** - Página de diagnóstico
5. **`/root/PulseExpends/DIAGNOSTICO-SOLUCION.md`** - Este documento

## 🔗 ENLACES DE PRUEBA

### Con frontend modificado (puertos directos):
- Frontend local: http://localhost:8889/
- MCP API: http://182.160.24.205:8080/health
- PDF API: http://182.160.24.205:8000/health

### Con proxy local:
- Aplicación completa: http://localhost:8888/
- APIs: http://localhost:8888/mcp/, http://localhost:8888/pdf/

### Servidor remoto (actual, mal configurado):
- Frontend: http://182.160.24.205/
- APIs (no funcionan vía nginx): http://182.160.24.205/mcp/, http://182.160.24.205/pdf/

## 🆘 PROBLEMA DE SSH

**Error**: `Permission denied (publickey,password)`

**Posibles causas**:
1. Clave SSH incorrecta o corrupta
2. Usuario incorrecto (no es `ubuntu`)
3. Servidor SSH configurado para no aceptar claves
4. Firewall bloqueando SSH
5. Instancia ECS sin configuración SSH

**Solución propuesta**:
1. Verificar la clave SSH: `ssh-keygen -y -f pulse-expends-key.pem`
2. Probar con usuario `root`: `ssh -i pulse-expends-key.pem root@182.160.24.205`
3. Verificar en Huawei Cloud Console la configuración SSH
4. Reiniciar la instancia ECS
5. Crear nueva clave SSH y asignarla a la instancia

## 📞 SOPORTE

**Contactar administrador del servidor** para:
1. Corregir configuración nginx en `/etc/nginx/sites-available/pulseexpends`
2. Verificar servicios systemd: `pulseexpends-mcp`, `pulseexpends-pdf-parser`, `pulseexpends-status`
3. Resolver problema de acceso SSH
4. Iniciar Status Dashboard en puerto 8081

**Comandos de diagnóstico remoto** (si se obtiene acceso SSH):
```bash
# Ver configuración nginx actual
sudo nginx -T
sudo cat /etc/nginx/sites-available/*

# Ver servicios
sudo systemctl status nginx pulseexpends-mcp pulseexpends-pdf-parser pulseexpends-status

# Ver logs
sudo journalctl -u nginx --no-pager -n 50
sudo journalctl -u pulseexpends-mcp --no-pager -n 50

# Probar rutas localmente
curl http://localhost/mcp/health
curl http://localhost/pdf/health
```

---

**Última actualización**: 2026-05-20  
**Estado**: Workaround implementado, corrección de nginx pendiente  
**Responsable**: Sistema de diagnóstico automático PulseExpends