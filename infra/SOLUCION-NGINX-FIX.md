# Solución: Problema de Routing en Nginx para PulseExpends

## Problema Identificado
La aplicación PulseExpends no funcionaba correctamente porque nginx estaba mal configurado:
- Las rutas `/mcp/*` y `/pdf/*` devolvían el frontend en lugar de redirigir a las APIs
- El frontend usaba rutas `/mcp` y `/pdf` pero nginx esperaba `/api/mcp/` y `/api/pdf/`
- La configuración `default_server` no estaba establecida correctamente

## Solución Aplicada
Se corrigió la configuración de nginx en el servidor ECS (`182.160.24.205`):

### Cambios realizados:
1. **Configuración de rutas corregida**: Se agregaron las rutas `/mcp/` y `/pdf/` además de `/api/mcp/` y `/api/pdf/` para compatibilidad
2. **default_server agregado**: Se configuró `listen 80 default_server` y `listen [::]:80 default_server`
3. **try_files corregido**: Se cambió de `try_files / /index.html;` a `try_files $uri $uri/ /index.html;`
4. **Sitio default deshabilitado**: Se eliminó el enlace simbólico a `/etc/nginx/sites-enabled/default`

### Configuración final (funcionando):
```nginx
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;

    root /opt/PulseExpends/services/frontend;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /mcp/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
        proxy_buffering off;
        proxy_set_header Connection '';
    }
    
    location /api/mcp/ {
        proxy_pass http://localhost:8080/;
        # ... misma configuración
    }

    location /pdf/ {
        proxy_pass http://localhost:8000/;
        # ... misma configuración
    }
    
    location /api/pdf/ {
        proxy_pass http://localhost:8000/;
        # ... misma configuración
    }
}
```

## Pasos para replicar la solución

### 1. Conectarse al servidor ECS
```bash
ssh -i pulse-expends-key.pem root@182.160.24.205
```

### 2. Corregir configuración de nginx
```bash
# Backup de la configuración actual
cp /etc/nginx/sites-available/pulseexpends /etc/nginx/sites-available/pulseexpends.backup.$(date +%Y%m%d_%H%M%S)

# Crear nueva configuración (ver archivo pulseexpends-correct.conf)
cat > /etc/nginx/sites-available/pulseexpends << 'EOF'
[configuración completa aquí]
EOF

# Asegurar que sea default_server
sed -i 's/listen 80;/listen 80 default_server;/' /etc/nginx/sites-available/pulseexpends
sed -i 's/listen \[::\]:80;/listen \[::\]:80 default_server;/' /etc/nginx/sites-available/pulseexpends

# Deshabilitar sitio default si existe
rm -f /etc/nginx/sites-enabled/default

# Verificar sintaxis
nginx -t

# Recargar nginx
systemctl reload nginx
```

### 3. Verificar que funcione
```bash
# Frontend
curl -I http://localhost/

# APIs
curl http://localhost/mcp/health
curl http://localhost/pdf/health
curl http://localhost/mcp/transactions
```

## Estado actual (verificado)
- ✅ Frontend: http://182.160.24.205/ (HTTP 200 OK)
- ✅ API MCP Server: http://182.160.24.205/mcp/health (`{"status":"healthy"}`)
- ✅ API PDF Parser: http://182.160.24.205/pdf/health (`{"status":"healthy"}`)
- ✅ Transacciones: http://182.160.24.205/mcp/transactions (3 transacciones)
- ✅ Resumen: http://182.160.24.205/mcp/summary (balance: $1,329.20)

## Lecciones aprendidas
1. **Siempre verificar rutas**: El frontend usaba `/mcp` pero nginx esperaba `/api/mcp/`
2. **default_server es crucial**: Sin él, nginx usa la configuración `default`
3. **try_files debe ser correcto**: `try_files $uri $uri/ /index.html;` no `try_files / /index.html;`
4. **Mantener compatibilidad**: Agregar ambas rutas (`/mcp/` y `/api/mcp/`) para evitar problemas futuros

## Archivos de referencia
- `infra/src/nginx/pulseexpends-correct.conf` - Configuración corregida
- `infra/src/nginx/pulseexpends.conf` - Configuración anterior (con problemas)

## Comandos útiles para diagnóstico
```bash
# Ver configuración nginx
nginx -T

# Ver archivos de configuración
ls -la /etc/nginx/sites-available/
ls -la /etc/nginx/sites-enabled/

# Ver logs de nginx
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log

# Ver servicios activos
systemctl status nginx pulseexpends-mcp-server pulseexpends-pdf-parser

# Probar endpoints
curl -I http://localhost/
curl http://localhost/mcp/health
curl http://localhost/pdf/health
```

## Para futuros despliegues
Incluir esta configuración corregida en los scripts de despliegue y verificar:
1. Que nginx tenga `default_server` configurado
2. Que las rutas coincidan entre frontend y backend
3. Que no haya conflicto con el sitio `default`
4. Probar todos los endpoints después del despliegue