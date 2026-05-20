#!/bin/bash

# Script de despliegue para configuración de subdominios PulseExpends
# Este script configura Nginx con subdominios para los diferentes servicios

set -e

echo "========================================="
echo "Despliegue de configuración de subdominios"
echo "========================================="

# Variables
DOMAIN="pulseexpends.duckdns.org"
SERVER_IP="182.160.24.205"
NGINX_CONF_DIR="/etc/nginx/sites-available"
NGINX_ENABLED_DIR="/etc/nginx/sites-enabled"
BACKUP_DIR="/etc/nginx/backup"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🔧 Configurando subdominios para:"
echo "   - Frontend: $DOMAIN"
echo "   - API: api.$DOMAIN"
echo "   - PDF Parser: pdf.$DOMAIN"
echo "   - Status: status.$DOMAIN"
echo "   - Servidor: $SERVER_IP"

# Verificar que estamos en el directorio correcto
if [ ! -f "backend/nginx-config/nginx-subdomains.conf" ]; then
    echo "❌ Error: No se encontró el archivo de configuración nginx-subdomains.conf"
    echo "   Ejecuta desde el directorio raíz del proyecto: /root/PulseExpends"
    exit 1
fi

# Crear backup de la configuración actual
echo "📦 Creando backup de configuración actual..."
sudo mkdir -p $BACKUP_DIR
if [ -f "$NGINX_CONF_DIR/pulseexpends" ]; then
    sudo cp $NGINX_CONF_DIR/pulseexpends $BACKUP_DIR/pulseexpends.backup.$TIMESTAMP
    echo "   ✅ Backup creado: $BACKUP_DIR/pulseexpends.backup.$TIMESTAMP"
fi

# Copiar nueva configuración de Nginx
echo "📝 Copiando nueva configuración de Nginx..."
sudo cp backend/nginx-config/nginx-subdomains.conf $NGINX_CONF_DIR/pulseexpends

# Crear enlace simbólico si no existe
if [ ! -L "$NGINX_ENABLED_DIR/pulseexpends" ]; then
    echo "🔗 Creando enlace simbólico..."
    sudo ln -s $NGINX_CONF_DIR/pulseexpends $NGINX_ENABLED_DIR/pulseexpends
fi

# Verificar sintaxis de Nginx
echo "🔍 Verificando sintaxis de Nginx..."
sudo nginx -t

if [ $? -eq 0 ]; then
    echo "✅ Sintaxis de Nginx OK"
    
    # Recargar Nginx
    echo "🔄 Recargando configuración de Nginx..."
    sudo systemctl reload nginx
    
    if [ $? -eq 0 ]; then
        echo "✅ Nginx recargado exitosamente"
        
        # Verificar servicios
        echo "🔍 Verificando servicios..."
        
        # Verificar MCP Server
        if systemctl is-active --quiet pulseexpends-mcp-server; then
            echo "✅ MCP Server está activo"
        else
            echo "⚠️  MCP Server no está activo. Iniciando..."
            sudo systemctl start pulseexpends-mcp-server
            sudo systemctl enable pulseexpends-mcp-server
        fi
        
        # Verificar PDF Parser
        if systemctl is-active --quiet pulseexpends-pdf-parser; then
            echo "✅ PDF Parser está activo"
        else
            echo "⚠️  PDF Parser no está activo. Iniciando..."
            sudo systemctl start pulseexpends-pdf-parser
            sudo systemctl enable pulseexpends-pdf-parser
        fi
        
        # Verificar Nginx
        if systemctl is-active --quiet nginx; then
            echo "✅ Nginx está activo"
        else
            echo "⚠️  Nginx no está activo. Iniciando..."
            sudo systemctl start nginx
            sudo systemctl enable nginx
        fi
        
        # Configurar DNS local para desarrollo (opcional)
        echo "🌐 Configurando DNS local para desarrollo..."
        if ! grep -q "pulseexpends.duckdns.org" /etc/hosts; then
            echo "# PulseExpends Development DNS" | sudo tee -a /etc/hosts
            echo "127.0.0.1 pulseexpends.duckdns.org" | sudo tee -a /etc/hosts
            echo "127.0.0.1 api.pulseexpends.duckdns.org" | sudo tee -a /etc/hosts
            echo "127.0.0.1 pdf.pulseexpends.duckdns.org" | sudo tee -a /etc/hosts
            echo "127.0.0.1 status.pulseexpends.duckdns.org" | sudo tee -a /etc/hosts
            echo "✅ DNS local configurado"
        else
            echo "ℹ️  DNS local ya configurado"
        fi
        
        # Mostrar resumen
        echo ""
        echo "========================================="
        echo "✅ DESPLIEGUE COMPLETADO EXITOSAMENTE"
        echo "========================================="
        echo ""
        echo "🌐 URLs de acceso:"
        echo "   Frontend Principal: http://$DOMAIN"
        echo "   API MCP Server:     http://api.$DOMAIN"
        echo "   PDF Parser:         http://pdf.$DOMAIN"
        echo "   Status Dashboard:   http://status.$DOMAIN"
        echo ""
        echo "🔧 Configuración de DNS:"
        echo "   Asegúrate de que estos registros apunten a $SERVER_IP:"
        echo "   - $DOMAIN"
        echo "   - api.$DOMAIN"
        echo "   - pdf.$DOMAIN"
        echo "   - status.$DOMAIN"
        echo ""
        echo "📋 Comandos útiles:"
        echo "   Ver logs de Nginx:      sudo journalctl -u nginx -f"
        echo "   Ver logs MCP Server:    sudo journalctl -u pulseexpends-mcp-server -f"
        echo "   Ver logs PDF Parser:    sudo journalctl -u pulseexpends-pdf-parser -f"
        echo "   Reiniciar servicios:    sudo systemctl restart nginx pulseexpends-mcp-server pulseexpends-pdf-parser"
        echo ""
        echo "⚠️  NOTA: El servidor ECS está actualmente APAGADO (SHUTOFF)"
        echo "   Inicia la instancia desde Huawei Cloud Console para que los servicios estén disponibles"
        
    else
        echo "❌ Error al recargar Nginx"
        echo "   Revertiendo cambios..."
        if [ -f "$BACKUP_DIR/pulseexpends.backup.$TIMESTAMP" ]; then
            sudo cp $BACKUP_DIR/pulseexpends.backup.$TIMESTAMP $NGINX_CONF_DIR/pulseexpends
            sudo systemctl reload nginx
            echo "   ✅ Configuración restaurada desde backup"
        fi
        exit 1
    fi
else
    echo "❌ Error en la sintaxis de Nginx"
    echo "   Revertiendo cambios..."
    if [ -f "$BACKUP_DIR/pulseexpends.backup.$TIMESTAMP" ]; then
        sudo cp $BACKUP_DIR/pulseexpends.backup.$TIMESTAMP $NGINX_CONF_DIR/pulseexpends
        echo "   ✅ Configuración restaurada desde backup"
    fi
    exit 1
fi

echo ""
echo "🎉 ¡Configuración completada! Los subdominios están listos para usar."
echo "   Recuerda iniciar la instancia ECS desde Huawei Cloud Console."