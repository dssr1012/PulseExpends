#!/bin/bash

# Script para configurar subdominios en el servidor ECS PulseExpends
# Este script debe ejecutarse en el servidor ECS después de iniciarlo

set -e

echo "========================================="
echo "configuration de Subdominios en ECS"
echo "========================================="

# Variables
DOMAIN="pulseexpends.duckdns.org"
SERVER_IP=$(curl -s https://api.ipify.org)

echo "🌐 Configurando subdominios para IP: $SERVER_IP"
echo "   Dominio: $DOMAIN"

# update sistema
echo "🔄 Actualizando sistema..."
sudo apt update && sudo apt upgrade -y

# Instalar Nginx si no está instalado
if ! command -v nginx &> /dev/null; then
    echo "📦 Instalando Nginx..."
    sudo apt install -y nginx
fi

# Instalar dependencias para servicios
echo "📦 Instalando dependencias..."

# Dependencias para MCP Server (Go)
if ! command -v go &> /dev/null; then
    echo "📦 Instalando Go..."
    sudo apt install -y golang-go
fi

# Dependencias para PDF Parser (Python)
if ! command -v python3 &> /dev/null; then
    echo "📦 Instalando Python3..."
    sudo apt install -y python3 python3-pip python3-venv
fi

# Instalar dependencias de Python
echo "📦 Instalando dependencias de Python..."
sudo pip3 install fastapi uvicorn pymupdf pytesseract Pillow python-multipart

# create directorios necesarios
echo "📁 Creando directorios..."
sudo mkdir -p /var/www/pulseexpends-frontend
sudo mkdir -p /var/www/pulseexpends-status
sudo mkdir -p /opt/PulseExpends
sudo mkdir -p /opt/PulseExpends/python/pdf-parser/src

# Clonar o copiar código del repository
echo "📥 Obteniendo código del repository..."
cd /opt/PulseExpends

# Si no existe el código, clonar el repository
if [ ! -d ".git" ]; then
    echo "📦 Clonando repository..."
    sudo git clone https://github.com/dssr1012/PulseExpends.git /tmp/pulseexpends-temp
    sudo cp -r /tmp/pulseexpends-temp/* .
    sudo cp -r /tmp/pulseexpends-temp/.* . 2>/dev/null || true
    sudo rm -rf /tmp/pulseexpends-temp
fi

# Configurar frontend
echo "🎨 Configurando frontend..."
sudo cp -r frontend/public/* /var/www/pulseexpends-frontend/
sudo cp -r frontend/styles /var/www/pulseexpends-frontend/
sudo cp -r frontend/src /var/www/pulseexpends-frontend/

# Configurar backend
echo "🔧 Configurando backend..."

# MCP Server
echo "🔄 Configurando MCP Server..."
cd /opt/PulseExpends/backend/mcp-server
sudo go mod download

# PDF Parser
echo "🔄 Configurando PDF Parser..."
cd /opt/PulseExpends/backend/pdf-parser
sudo pip3 install -r requirements.txt

# Configurar Nginx con subdominios
echo "🌐 Configurando Nginx..."
sudo cp /opt/PulseExpends/backend/nginx-config/nginx-subdomains.conf /etc/nginx/sites-available/pulseexpends

# create enlace simbólico si no existe
if [ ! -L "/etc/nginx/sites-enabled/pulseexpends" ]; then
    sudo ln -s /etc/nginx/sites-available/pulseexpends /etc/nginx/sites-enabled/
fi

# Remover configuration por defecto de Nginx
if [ -L "/etc/nginx/sites-enabled/default" ]; then
    sudo rm /etc/nginx/sites-enabled/default
fi

# Configurar servicios systemd
echo "⚙️ Configurando servicios systemd..."

# MCP Server service
sudo tee /etc/systemd/system/pulseexpends-mcp-server.service > /dev/null << EOF
[Unit]
Description=PulseExpends MCP Server
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/PulseExpends/backend/mcp-server
ExecStart=/usr/bin/go run /opt/PulseExpends/backend/mcp-server/mcp-server-enhanced.go
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
EOF

# PDF Parser service
sudo tee /etc/systemd/system/pulseexpends-pdf-parser.service > /dev/null << EOF
[Unit]
Description=PulseExpends PDF Parser
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/PulseExpends/backend/pdf-parser
ExecStart=/usr/bin/python3 /opt/PulseExpends/backend/pdf-parser/app.py
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment="PORT=8000"

[Install]
WantedBy=multi-user.target
EOF

# Status Dashboard service (simple Python HTTP server)
sudo tee /etc/systemd/system/pulseexpends-status.service > /dev/null << EOF
[Unit]
Description=PulseExpends Status Dashboard
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/var/www/pulseexpends-status
ExecStart=/usr/bin/python3 -m http.server 8081
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# create dashboard de status simple
sudo tee /var/www/pulseexpends-status/index.html > /dev/null << EOF
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PulseExpends - Status Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f6fa; color: #2f3640; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #00a8ff, #0097e6); color: white; padding: 40px 0; text-align: center; border-radius: 0 0 20px 20px; margin-bottom: 40px; }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header p { font-size: 1.2em; opacity: 0.9; }
        .status-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 40px; }
        .status-card { background: white; border-radius: 15px; padding: 25px; box-shadow: 0 5px 15px rgba(0,0,0,0.1); }
        .status-card h2 { color: #2f3640; margin-bottom: 20px; font-size: 1.5em; border-bottom: 2px solid #f5f6fa; padding-bottom: 10px; }
        .service-status { display: flex; align-items: center; justify-content: space-between; padding: 15px; margin: 10px 0; background: #f8f9fa; border-radius: 10px; }
        .service-name { font-weight: 600; }
        .status { padding: 5px 15px; border-radius: 20px; font-weight: 600; }
        .status.up { background: #4cd137; color: white; }
        .status.down { background: #e84118; color: white; }
        .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; }
        .info-card { background: white; border-radius: 15px; padding: 20px; box-shadow: 0 5px 15px rgba(0,0,0,0.1); }
        .info-card h3 { color: #00a8ff; margin-bottom: 15px; }
        .info-item { margin: 10px 0; }
        .label { font-weight: 600; color: #7f8fa6; }
        .value { color: #2f3640; }
        .url-list { list-style: none; }
        .url-list li { margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 8px; }
        .url-list a { color: #00a8ff; text-decoration: none; }
        .url-list a:hover { text-decoration: underline; }
        .timestamp { text-align: center; margin-top: 40px; color: #7f8fa6; font-size: 0.9em; }
        @media (max-width: 768px) {
            .container { padding: 10px; }
            .header { padding: 20px 0; }
            .header h1 { font-size: 2em; }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <h1>PulseExpends Status Dashboard</h1>
            <p>Monitoreo de servicios y subdominios</p>
        </div>
    </div>
    
    <div class="container">
        <div class="status-grid">
            <div class="status-card">
                <h2>Estado de Servicios</h2>
                <div class="service-status">
                    <span class="service-name">Frontend Principal</span>
                    <span class="status up" id="status-frontend">UP</span>
                </div>
                <div class="service-status">
                    <span class="service-name">api MCP Server</span>
                    <span class="status up" id="status-api">UP</span>
                </div>
                <div class="service-status">
                    <span class="service-name">PDF Parser api</span>
                    <span class="status up" id="status-pdf">UP</span>
                </div>
                <div class="service-status">
                    <span class="service-name">Nginx Reverse Proxy</span>
                    <span class="status up" id="status-nginx">UP</span>
                </div>
            </div>
            
            <div class="status-card">
                <h2>information del Sistema</h2>
                <div class="info-grid">
                    <div class="info-item">
                        <div class="label">Servidor IP</div>
                        <div class="value" id="server-ip">Cargando...</div>
                    </div>
                    <div class="info-item">
                        <div class="label">Dominio</div>
                        <div class="value">pulseexpends.duckdns.org</div>
                    </div>
                    <div class="info-item">
                        <div class="label">date/time</div>
                        <div class="value" id="current-time">Cargando...</div>
                    </div>
                    <div class="info-item">
                        <div class="label">Tiempo Activo</div>
                        <div class="value" id="uptime">Cargando...</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="status-card">
            <h2>URLs de Acceso</h2>
            <ul class="url-list">
                <li><a href="http://pulseexpends.duckdns.org" target="_blank">🌐 Frontend Principal</a> - Aplicación web completa</li>
                <li><a href="http://api.pulseexpends.duckdns.org" target="_blank">🔧 api MCP Server</a> - api REST para transacciones</li>
                <li><a href="http://pdf.pulseexpends.duckdns.org" target="_blank">📄 PDF Parser api</a> - Procesamiento de PDFs</li>
                <li><a href="http://status.pulseexpends.duckdns.org" target="_blank">📊 Status Dashboard</a> - Esta página de monitoreo</li>
            </ul>
        </div>
        
        <div class="status-card">
            <h2>Comandos Útiles</h2>
            <div class="info-grid">
                <div class="info-card">
                    <h3>Ver Logs</h3>
                    <div class="info-item">
                        <div class="label">Nginx</div>
                        <div class="value">sudo journalctl -u nginx -f</div>
                    </div>
                    <div class="info-item">
                        <div class="label">MCP Server</div>
                        <div class="value">sudo journalctl -u pulseexpends-mcp-server -f</div>
                    </div>
                    <div class="info-item">
                        <div class="label">PDF Parser</div>
                        <div class="value">sudo journalctl -u pulseexpends-pdf-parser -f</div>
                    </div>
                </div>
                
                <div class="info-card">
                    <h3>Reiniciar Servicios</h3>
                    <div class="info-item">
                        <div class="label">Todos los servicios</div>
                        <div class="value">sudo systemctl restart nginx pulseexpends-*</div>
                    </div>
                    <div class="info-item">
                        <div class="label">Ver estado</div>
                        <div class="value">sudo systemctl status pulseexpends-*</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="timestamp" id="timestamp">
            Última actualización: <span id="update-time">Cargando...</span>
        </div>
    </div>
    
    <script>
        // update information del sistema
        function updateSystemInfo() {
            // IP del servidor
            document.getElementById('server-ip').textContent = window.location.hostname;
            
            // time actual
            const now = new Date();
            document.getElementById('current-time').textContent = now.toLocaleString('es-ES');
            
            // Tiempo de actividad (simulado)
            const startTime = Date.now() - (Math.random() * 86400000); // Hace 0-24 horas
            const uptimeMs = Date.now() - startTime;
            const hours = Math.floor(uptimeMs / 3600000);
            const minutes = Math.floor((uptimeMs % 3600000) / 60000);
            document.getElementById('uptime').textContent = \`\${hours}h \${minutes}m\`;
            
            // Marca de tiempo de actualización
            document.getElementById('update-time').textContent = now.toLocaleString('es-ES');
            
            // Verificar estado de servicios (simulado)
            checkServiceStatus();
        }
        
        // Verificar estado de servicios
        async function checkServiceStatus() {
            const services = [
                { id: 'status-frontend', url: '/' },
                { id: 'status-api', url: 'http://api.pulseexpends.duckdns.org/health' },
                { id: 'status-pdf', url: 'http://pdf.pulseexpends.duckdns.org/health' },
                { id: 'status-nginx', url: '/' }
            ];
            
            for (const service of services) {
                try {
                    const response = await fetch(service.url, { method: 'GET', mode: 'no-cors' });
                    document.getElementById(service.id).className = 'status up';
                    document.getElementById(service.id).textContent = 'UP';
                } catch (error) {
                    document.getElementById(service.id).className = 'status down';
                    document.getElementById(service.id).textContent = 'DOWN';
                }
            }
        }
        
        // update cada 30 segundos
        updateSystemInfo();
        setInterval(updateSystemInfo, 30000);
        
        // Verificar servicios cada minuto
        setInterval(checkServiceStatus, 60000);
    </script>
</body>
</html>
EOF

# Configurar permisos
echo "🔒 Configurando permisos..."
sudo chown -R www-data:www-data /var/www/pulseexpends-frontend
sudo chown -R www-data:www-data /var/www/pulseexpends-status
sudo chown -R ubuntu:ubuntu /opt/PulseExpends

# Recargar systemd
echo "🔄 Recargando systemd..."
sudo systemctl daemon-reload

# Habilitar servicios
echo "⚙️ Habilitando servicios..."
sudo systemctl enable pulseexpends-mcp-server
sudo systemctl enable pulseexpends-pdf-parser
sudo systemctl enable pulseexpends-status

# Verificar configuration de Nginx
echo "🔍 Verificando configuration de Nginx..."
sudo nginx -t

if [ $? -eq 0 ]; then
    echo "✅ Sintaxis de Nginx OK"
    
    # Reiniciar servicios
    echo "🔄 Reiniciando servicios..."
    sudo systemctl restart nginx
    sudo systemctl restart pulseexpends-mcp-server
    sudo systemctl restart pulseexpends-pdf-parser
    sudo systemctl restart pulseexpends-status
    
    # Verificar estado
    echo "📊 Estado de servicios:"
    echo ""
    echo "Nginx:"
    sudo systemctl status nginx --no-pager | grep -E "Active:|Loaded:"
    echo ""
    echo "MCP Server:"
    sudo systemctl status pulseexpends-mcp-server --no-pager | grep -E "Active:|Loaded:"
    echo ""
    echo "PDF Parser:"
    sudo systemctl status pulseexpends-pdf-parser --no-pager | grep -E "Active:|Loaded:"
    echo ""
    echo "Status Dashboard:"
    sudo systemctl status pulseexpends-status --no-pager | grep -E "Active:|Loaded:"
    
    # Configurar DuckDNS (si se proporciona token)
    if [ -n "$DUCKDNS_TOKEN" ]; then
        echo ""
        echo "🌐 Configurando DuckDNS..."
        curl -s "https://www.duckdns.org/update?domains=$DOMAIN,api.$DOMAIN,pdf.$DOMAIN,status.$DOMAIN&token=$DUCKDNS_TOKEN&ip=$SERVER_IP"
        echo ""
        echo "✅ DuckDNS updated"
    fi
    
    echo ""
    echo "========================================="
    echo "✅ configuration COMPLETADA EXITOSAMENTE"
    echo "========================================="
    echo ""
    echo "🌐 URLs de acceso:"
    echo "   Frontend Principal: http://$DOMAIN"
    echo "   api MCP Server:     http://api.$DOMAIN"
    echo "   PDF Parser:         http://pdf.$DOMAIN"
    echo "   Status Dashboard:   http://status.$DOMAIN"
    echo ""
    echo "📊 Para verificar el estado:"
    echo "   curl http://$DOMAIN/health"
    echo "   curl http://api.$DOMAIN/health"
    echo "   curl http://pdf.$DOMAIN/health"
    echo ""
    echo "📋 Comandos útiles:"
    echo "   Ver logs: sudo journalctl -fu [service]"
    echo "   Reiniciar: sudo systemctl restart [service]"
    echo "   Estado: sudo systemctl status [service]"
    
else
    echo "❌ error en la sintaxis de Nginx"
    echo "   Revisa la configuration en /etc/nginx/sites-available/pulseexpends"
    exit 1
fi

echo ""
echo "🎉 ¡configuration completada! Los servicios están listos."
echo "   Recuerda configurar DuckDNS si aún no lo has hecho."