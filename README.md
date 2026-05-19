# PulseExpends - Sistema de Gestión de Gastos

PulseExpends es un sistema completo de gestión de gastos inspirado en Fintonic, que permite a los usuarios gestionar sus finanzas personales, subir resúmenes de tarjetas de crédito en PDF y visualizar análisis financieros.

## 🚀 Características Principales

### Frontend (Aplicación Web)
- **Dashboard interactivo** con estadísticas en tiempo real
- **Gestión de transacciones** (agregar, editar, eliminar)
- **Subida de PDFs** con drag-and-drop
- **Análisis por categorías** con gráficos visuales
- **Sistema de alertas** para gastos altos
- **Diseño responsivo** mobile-first
- **Interfaz en español** inspirada en Fintonic

### Backend (APIs)
- **Servidor MCP** (Go) - API REST para gestión de transacciones
- **PDF Parser** (Python/FastAPI) - Extracción de transacciones de PDFs
- **Nginx** - Reverse proxy unificado
- **Systemd** - Gestión de servicios como demonios

## 🏗️ Arquitectura

```
PulseExpends/
├── frontend/              # Aplicación web completa
│   ├── public/           # Archivos estáticos (HTML)
│   │   └── index.html    # Página principal SPA
│   ├── src/              # Código fuente JavaScript
│   │   └── app.js        # Lógica principal de la aplicación
│   └── styles/           # Estilos CSS
│       └── main.css      # Estilos principales
├── backend/              # Servicios backend
│   ├── mcp-server/       # Servidor MCP (Go)
│   ├── pdf-parser/       # Parser de PDFs (Python)
│   ├── nginx-config/     # Configuración de Nginx
│   └── systemd-services/ # Archivos de servicio systemd
└── README.md             # Este archivo
```

## 🛠️ Requisitos del Sistema

- **Servidor**: Ubuntu 20.04+ (o similar)
- **Nginx**: Versión 1.18+
- **Go**: Versión 1.21+ (para MCP Server)
- **Python**: Versión 3.8+ (para PDF Parser)
- **Node.js**: Opcional para desarrollo frontend

## 🌐 Configuración de Subdominios

PulseExpends utiliza subdominios para organizar los diferentes servicios:

- **Frontend Principal**: `http://pulseexpends.duckdns.org`
- **API MCP Server**: `http://api.pulseexpends.duckdns.org`
- **PDF Parser API**: `http://pdf.pulseexpends.duckdns.org`
- **Status Dashboard**: `http://status.pulseexpends.duckdns.org`

### Configuración DNS

1. **Configurar DuckDNS**:
   ```bash
   # Editar el script con tu token de DuckDNS
   nano duckdns-config.sh
   
   # Ejecutar configuración
   ./duckdns-config.sh
   ```

2. **Configurar Nginx con subdominios**:
   ```bash
   ./deploy-subdomains.sh
   ```

### Desarrollo Local

Para desarrollo local, puedes configurar `/etc/hosts`:
```
127.0.0.1 pulseexpends.duckdns.org
127.0.0.1 api.pulseexpends.duckdns.org
127.0.0.1 pdf.pulseexpends.duckdns.org
127.0.0.1 status.pulseexpends.duckdns.org
```

## 🚀 Despliegue Rápido

### 1. Configurar servidor
```bash
# Instalar dependencias
sudo apt update
sudo apt install -y nginx golang python3 python3-pip

# Instalar dependencias de Python para PDF Parser
pip3 install fastapi uvicorn pymupdf pytesseract Pillow python-multipart
```

### 2. Configurar Nginx
```bash
sudo cp backend/nginx-config/nginx.conf /etc/nginx/sites-available/pulseexpends
sudo ln -s /etc/nginx/sites-available/pulseexpends /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 3. Configurar servicios systemd
```bash
sudo cp backend/systemd-services/*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable pulseexpends-mcp-server
sudo systemctl enable pulseexpends-pdf-parser
sudo systemctl start pulseexpends-mcp-server
sudo systemctl start pulseexpends-pdf-parser
```

### 4. Desplegar frontend
```bash
sudo mkdir -p /var/www/pulseexpends-frontend
sudo cp -r frontend/* /var/www/pulseexpends-frontend/
sudo chown -R www-data:www-data /var/www/pulseexpends-frontend
```

### 5. Desplegar backend
```bash
sudo mkdir -p /opt/PulseExpends
sudo cp -r backend/mcp-server/* /opt/PulseExpends/
sudo cp -r backend/pdf-parser/* /opt/PulseExpends/python/pdf-parser/src/

# Instalar dependencias de Go
cd /opt/PulseExpends
go mod download
```

## 🔧 Configuración de Servicios

### Servidor MCP
- **Puerto**: 8080
- **Endpoints**:
  - `GET /transactions` - Listar transacciones
  - `POST /transactions` - Agregar transacción
  - `GET /summary` - Resumen por categoría
  - `GET /health` - Health check

### PDF Parser
- **Puerto**: 8000
- **Endpoints**:
  - `POST /parse` - Procesar archivo PDF
  - `GET /health` - Health check

### Nginx (Reverse Proxy)
- **Frontend**: `http://dominio.com/`
- **MCP API**: `http://dominio.com/api/mcp/`
- **PDF API**: `http://dominio.com/api/pdf/`

## 📁 Estructura de Datos

### Transacción
```json
{
  "id": "uuid",
  "amount": 75.50,
  "currency": "USD",
  "category": "Entertainment",
  "description": "Movie tickets",
  "type": "expense",
  "date": "2024-05-19T00:00:00Z"
}
```

### Respuesta del PDF Parser
```json
{
  "success": true,
  "message": "PDF procesado exitosamente",
  "data": {
    "text": "Texto extraído del PDF...",
    "transaction_count": 5
  }
}
```

## 🎨 Diseño Frontend

### Paleta de Colores (inspirada en Fintonic)
- **Azul principal**: `#00a8ff`
- **Azul oscuro**: `#0097e6`
- **Verde éxito**: `#4cd137`
- **Rojo error**: `#e84118`
- **Amarillo advertencia**: `#fbc531`
- **Gris claro**: `#f5f6fa`
- **Gris oscuro**: `#7f8fa6`

### Características de UI/UX
- **Navegación SPA** sin recargas de página
- **Modales** para agregar gastos y subir PDFs
- **Drag-and-drop** para subida de archivos
- **Filtros y búsqueda** en lista de transacciones
- **Gráficos de categorías** visuales
- **Diseño responsive** para móviles y desktop

## 🔒 Seguridad

- **Nginx** como reverse proxy con configuración segura
- **CORS** configurado para el servidor MCP
- **Límite de tamaño** de archivos PDF (10MB)
- **Validación** de tipos de archivo
- **Systemd** para gestión segura de servicios

## 📊 Monitoreo

### Health Checks
- MCP Server: `http://localhost:8080/health`
- PDF Parser: `http://localhost:8000/health`
- Nginx: `http://localhost/health`

### Logs de Systemd
```bash
# Ver logs del servidor MCP
sudo journalctl -u pulseexpends-mcp-server -f

# Ver logs del PDF Parser
sudo journalctl -u pulseexpends-pdf-parser -f

# Ver logs de Nginx
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## 🐛 Solución de Problemas

### Servidor MCP no inicia
```bash
# Verificar si el puerto está en uso
sudo netstat -tlnp | grep :8080

# Verificar logs
sudo journalctl -u pulseexpends-mcp-server --no-pager -n 50
```

### PDF Parser no procesa archivos
```bash
# Verificar dependencias de Python
pip3 list | grep -E "fastapi|uvicorn|pymupdf"

# Verificar permisos de archivos
ls -la /opt/PulseExpends/python/pdf-parser/
```

### Nginx no sirve el frontend
```bash
# Verificar configuración
sudo nginx -t

# Verificar permisos
ls -la /var/www/pulseexpends-frontend/

# Recargar configuración
sudo systemctl reload nginx
```

## 📈 Próximas Mejoras

1. **Autenticación de usuarios** (login/registro)
2. **Base de datos persistente** (PostgreSQL)
3. **Notificaciones push/email** reales
4. **Integración con APIs bancarias**
5. **Aplicación móvil** (PWA o nativa)
6. **Exportación de datos** (CSV, Excel, PDF)
7. **Análisis predictivo** de gastos
8. **Presupuestos personalizados** por categoría

## 📄 Licencia

Este proyecto está bajo la licencia MIT. Ver el archivo LICENSE para más detalles.

## 🤝 Contribuciones

Las contribuciones son bienvenidas. Por favor, abre un issue o pull request en GitHub.

## 🚀 Scripts de Despliegue

### 1. Configurar DuckDNS
```bash
# Primero, edita el script con tu token de DuckDNS
export DUCKDNS_TOKEN="tu_token_aqui"
sed -i "s/TU_TOKEN_DE_DUCKDNS_AQUI/$DUCKDNS_TOKEN/" duckdns-config.sh

# Luego ejecuta
./duckdns-config.sh
```

### 2. Desplegar con Subdominios
```bash
./deploy-subdomains.sh
```

### 3. Verificar Servicios
```bash
# Ver estado de todos los servicios
sudo systemctl status nginx pulseexpends-mcp-server pulseexpends-pdf-parser

# Ver logs en tiempo real
sudo journalctl -fu nginx
sudo journalctl -fu pulseexpends-mcp-server
sudo journalctl -fu pulseexpends-pdf-parser
```

### 4. Actualización Manual de DNS
```bash
# Script de actualización manual
sudo /usr/local/bin/update-duckdns.sh
```

## 📞 Soporte

Para soporte o preguntas, abre un issue en el repositorio o contacta al equipo de desarrollo.