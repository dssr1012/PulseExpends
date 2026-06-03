# PulseExpends Infrastructure & Frontend

## 🚀 Sistema Completo de Gestión de Gastos - Estilo Fintonic

**URL de Producción:** http://182.160.24.205/

## 📋 Descripción del Proyecto

PulseExpends es un sistema completo de gestión de gastos inspirado en Fintonic, desplegado en Huawei Cloud con un frontend web moderno y APIs RESTful.

### ✨ Características Principales

#### **Frontend Web (Estilo Fintonic)**
- 📊 Dashboard interactivo con estadísticas en tiempo real
- 💰 Gestión completa de gastos e ingresos
- 📄 Subida de resúmenes de tarjeta en PDF
- 🔔 Sistema de alertas y notificationes
- 📱 Diseño 100% responsivo (mobile, tablet, desktop)
- 🎨 Interfaz moderna inspirada en Fintonic

#### **Backend APIs**
- **MCP Server (Go)**: api REST para gestión de transactiones
- **PDF Parser (Python/Flask)**: Procesamiento de archivos PDF
- **Nginx**: Reverse proxy y servidor web

#### **Infraestructura (Huawei Cloud)**
- **ECS Instance**: `ac8.large.2` (2 vCPU, 8GB RAM)
- **Public IP**: `182.160.24.205`
- **VPC/Subnet**: Configurado con grupos de seguridad
- **EIP**: Facturación por tráfico (optimizado para costos)
- **OBS Buckets**: 2 buckets con encriptación KMS

## 🏗️ Estructura del Proyecto

```
PulseExpends-Infra/
├── frontend/                    # Aplicación web completa
│   ├── index.html              # Frontend principal (HTML/CSS/JS)
│   ├── mcp-server-enhanced.go  # Servidor MCP mejorado (Go)
│   ├── nginx.conf              # configuration de Nginx
│   └── README.md               # Documentación del frontend
├── dashboard/                  # Dashboard de infraestructura
│   ├── dashboard.html          # Dashboard HTML
│   ├── status.html            # Página de estado
│   ├── check-status.py        # Script de verification
│   ├── serve-dashboard.py     # Servidor Python del dashboard
│   └── README.md              # Documentación del dashboard
├── scripts/                    # Scripts de despliegue
│   ├── start-dashboard.sh     # Iniciar dashboard local
│   ├── deploy-dashboard-to-ecs.sh  # Desplegar a ECS
│   ├── deploy-with-password.sh     # Despliegue con password
│   ├── check-ecs-status.py    # Verificar estado de ECS
│   ├── update-nginx-config.py # update configuration Nginx
│   └── infrastructure-status.json  # Estado de infraestructura
├── backup/                    # configuration modular Terraform
├── main.tf                   # configuration principal Terraform
├── terraform.tfvars          # Variables Terraform
├── terraform.tfstate         # Estado de Terraform
├── DEPLOYMENT_SUMMARY.md     # Resumen de despliegue
└── README.md                 # Este archivo
```

## 🚀 Despliegue Rápido

### **Requisitos Previos**
1. Huawei Cloud Account
2. Terraform instalado
3. SSH key pair
4. Dominio (opcional, se usa DuckDNS)

### **Paso 1: Configurar Infraestructura**
```bash
# Inicializar Terraform
terraform init

# Planear despliegue
terraform plan

# Aplicar configuration
terraform apply -auto-approve
```

### **Paso 2: Desplegar Aplicación**
```bash
# Desplegar dashboard y frontend
./scripts/deploy-dashboard-to-ecs.sh

# O usar password SSH
./scripts/deploy-with-password.sh
```

### **Paso 3: Acceder a la Aplicación**
- **Frontend**: http://182.160.24.205/
- **Dashboard**: http://182.160.24.205/dashboard/
- **MCP api**: http://182.160.24.205/mcp/
- **PDF api**: http://182.160.24.205/pdf/

## 🔧 configuration de Servicios

### **Servicios Systemd**
```bash
# Ver estado de todos los servicios
systemctl status pulseexpends-*

# Servicios individuales
systemctl status pulseexpends-dashboard
systemctl status pulseexpends-mcp
systemctl status pulseexpends-pdf-parser
systemctl status nginx
```

### **configuration Nginx**
```nginx
# /etc/nginx/sites-available/pulseexpends
server {
    listen 80 default_server;
    
    # Frontend
    location / {
        root /var/www/pulseexpends-frontend;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # MCP Server api
    location /mcp/ {
        proxy_pass http://localhost:8080/;
    }
    
    # PDF Parser api
    location /pdf/ {
        proxy_pass http://localhost:8000/;
    }
}
```

## 💻 Desarrollo Frontend

### **Estructura del Frontend**
El frontend está construido con HTML/CSS/JavaScript puro (sin frameworks) para máxima simplicidad y rendimiento.

#### **Características Técnicas:**
- **Vanilla JavaScript**: Sin dependencias externas
- **CSS Grid/Flexbox**: Layouts responsivos
- **Font Awesome**: Iconos modernos
- **Fetch api**: Comunicación REST
- **Drag & Drop**: Para subida de PDFs
- **Local Storage**: Para preferencias de user

#### **APIs Consumidas:**
```javascript
// MCP Server - Gestión de transactiones
GET  /mcp/transactions     // list transactiones
POST /mcp/transactions     // Agregar transaction
GET  /mcp/summary          // Resumen por categoría

// PDF Parser - Procesamiento de PDFs
POST /pdf/api/parse        // Procesar archivo PDF
```

### **Ejemplos de Uso**

#### **Agregar Gasto desde Frontend**
```javascript
// Ejemplo: Agregar un gasto
const expenseData = {
    description: "Cena en restaurante",
    amount: 45.50,
    category: "Food",
    type: "expense",
    currency: "USD",
    date: new Date().toISOString()
};

fetch('/mcp/transactions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(expenseData)
});
```

#### **Subir PDF desde Frontend**
```javascript
// Ejemplo: Subir resumen de tarjeta
const formData = new FormData();
formData.append('file', pdfFile);

fetch('/pdf/api/parse', {
    method: 'POST',
    body: formData
});
```

## 📊 APIs Disponibles

### **MCP Server (Go) - Puerto 8080**
```bash
# Health check
curl http://localhost:8080/health

# list transactiones
curl http://localhost:8080/transactions

# Agregar transaction
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"description":"Supermercado","amount":85.50,"category":"Food","type":"expense"}'

# Resumen de gastos
curl http://localhost:8080/summary
```

### **PDF Parser (Python) - Puerto 8000**
```bash
# Health check
curl http://localhost:8000/health

# Procesar PDF
curl -X POST http://localhost:8000/api/parse \
  -F "file=@resumen.pdf"
```

## 🎨 Diseño UI/UX

### **Inspiración Fintonic**
- **Paleta de colores**: Azules y verdes corporativos
- **Tarjetas de estadísticas**: Con iconos y tendencias
- **Lista de transactiones**: Con categorías visuales
- **Gráficos circulares**: Para distribución de gastos
- **Sistema de alertas**: Notificaciones prominentes

### **Componentes Principales**
1. **Dashboard**: view general con KPI principales
2. **Formulario de gastos**: Entrada simple con validation
3. **Lista de transactiones**: Filtrable y buscable
4. **Subida de PDF**: Drag & drop con feedback visual
5. **Análisis por categoría**: Gráficos y porcentajes
6. **Panel de alertas**: configuration y notificationes

## 🔒 Seguridad

### **Configuraciones Implementadas**
- **Security Groups**: Solo puertos necesarios abiertos (22, 80, 443, 8080, 8000)
- **EBS Encryption**: Volúmenes encriptados
- **OBS Encryption**: Buckets con KMS
- **SSH Key Authentication**: Sin passwords por defecto
- **Nginx como Reverse Proxy**: Protección adicional

### **Recomendaciones para Producción**
1. **Habilitar HTTPS** con certificados SSL
2. **Implementar authentication** de users
3. **Configurar WAF** (Web Application Firewall)
4. **Habilitar logging** y monitoreo
5. **Realizar backups** regulares de la base de datos

## 📈 Monitoreo y Mantenimiento

### **Dashboard de Infraestructura**
Accesible en: http://182.160.24.205/dashboard/
- Estado de servicios
- Uso de recursos
- Logs del sistema
- Métricas de rendimiento

### **Comandos de Mantenimiento**
```bash
# Ver logs de servicios
journalctl -u pulseexpends-dashboard -f
journalctl -u pulseexpends-mcp -f
journalctl -u pulseexpends-pdf-parser -f

# Reiniciar servicios
systemctl restart pulseexpends-dashboard
systemctl restart pulseexpends-mcp
systemctl restart pulseexpends-pdf-parser
systemctl restart nginx

# Ver uso de recursos
htop
df -h
free -h
```

## 🔄 Actualizaciones

### **update Frontend**
```bash
# Copiar nuevos archivos al servidor
scp -i pulse-expends-key.pem frontend/* ubuntu@182.160.24.205:/var/www/pulseexpends-frontend/

# Reiniciar Nginx
ssh -i pulse-expends-key.pem ubuntu@182.160.24.205 "sudo systemctl restart nginx"
```

### **update Backend**
```bash
# Recompilar y desplegar MCP Server
cd frontend
go build -o mcp-server-enhanced mcp-server-enhanced.go
scp -i pulse-expends-key.pem mcp-server-enhanced ubuntu@182.160.24.205:/opt/

# Reiniciar service
ssh -i pulse-expends-key.pem ubuntu@182.160.24.205 "sudo systemctl restart pulseexpends-mcp"
```

## 🐛 Solución de Problemas

### **Frontend no carga**
```bash
# Verificar Nginx
ssh root@182.160.24.205 "systemctl status nginx"
ssh root@182.160.24.205 "nginx -t"

# Verificar archivos
ssh root@182.160.24.205 "ls -la /var/www/pulseexpends-frontend/"
```

### **APIs no responden**
```bash
# Verificar servicios
ssh root@182.160.24.205 "systemctl status pulseexpends-mcp pulseexpends-pdf-parser"

# Probar endpoints
curl http://182.160.24.205/mcp/health
curl http://182.160.24.205/pdf/health

# Ver logs
ssh root@182.160.24.205 "journalctl -u pulseexpends-mcp --no-pager -l"
```

### **Problemas con PDF upload**
1. Verificar que el archivo sea PDF (no escaneado/imagen)
2. Tamaño máximo: 10MB
3. El PDF debe contener texto (no solo imágenes)
4. Revisar logs del PDF Parser

## 📞 Soporte

### **Recursos**
- **Documentación**: Este README y archivos en `/docs/`
- **Dashboard**: http://182.160.24.205/dashboard/
- **api Docs**: Endpoints documentados en `frontend/README.md`

### **Contacto**
- **Issues**: Reportar problemas en GitHub
- **Pull Requests**: Contribuciones bienvenidas
- **Discusiones**: Para preguntas y sugerencias

## 📄 Licencia

Este proyecto está bajo la licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 🙏 Agradecimientos

- **Huawei Cloud** por la infraestructura
- **Fintonic** por la inspiración del diseño
- **Open Source Community** por las herramientas utilizadas

## 🗄️ Base de Datos RDS PostgreSQL

### **Características de RDS:**
- **PostgreSQL 15** con alta disponibilidad
- **Backups automáticos** con retención de 7 días
- **Encriptación en reposo** con KMS
- **Monitoreo integrado** con Cloud Eye
- **Escalabilidad vertical** según demanda

### **configuration:**
```bash
# Habilitar RDS en terraform.tfvars
enable_rds = true
rds_instance_type = "rds.pg.c2.medium"
rds_storage = 100
rds_username = "pulseexpends_admin"
rds_password = "TuContraseñaSegura123!"

# Aplicar configuration
terraform apply -var-file="rds.auto.tfvars"
```

### **Conectar Aplicación a RDS:**
```bash
# update DATABASE_URL en .env
DATABASE_URL=postgresql://pulseexpends_admin:password@rds-endpoint:5432/pulseexpends

# Reiniciar servicios
echo "DATABASE_URL=postgresql://pulseexpends_admin:password@rds-endpoint:5432/pulseexpends" >> /opt/PulseExpends/backend/auth/.env
sudo systemctl restart pulseexpends-auth.service
```

### **Documentación Completa:**
Ver [RDS-README.md](RDS-README.md) para detalles completos de configuration, operation y mantenimiento.

---

**¡Sistema listo para producción!** 🚀

**URL Principal:** http://182.160.24.205/
**Dashboard:** http://182.160.24.205/dashboard/
**Estado:** Todos los servicios functionando ✅

Para comenzar a usar el sistema, simplemente visita la URL principal y comienza a agregar gastos o subir resúmenes de tarjeta.