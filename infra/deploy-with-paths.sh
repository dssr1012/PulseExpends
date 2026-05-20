#!/bin/bash

# Script de despliegue para configuración con paths (sin subdominios)
# Versión estable que funciona con DuckDNS

set -e

echo "========================================="
echo "Despliegue de PulseExpends con Paths"
echo "========================================="

# Variables
DOMAIN="pulseexpends.duckdns.org"
SERVER_IP="182.160.24.205"
NGINX_CONF_DIR="/etc/nginx/sites-available"
NGINX_ENABLED_DIR="/etc/nginx/sites-enabled"
BACKUP_DIR="/etc/nginx/backup"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🔧 Configurando PulseExpends con paths:"
echo "   - Dominio: $DOMAIN"
echo "   - IP del servidor: $SERVER_IP"
echo ""
echo "🌐 URLs de acceso:"
echo "   - Frontend Principal: http://$DOMAIN/"
echo "   - API MCP Server:     http://$DOMAIN/mcp/"
echo "   - PDF Parser:         http://$DOMAIN/pdf/"
echo "   - Status Dashboard:   http://$DOMAIN/status/"
echo ""

# Verificar que estamos en el directorio correcto
if [ ! -f "infra/src/nginx/pulseexpends.conf" ]; then
    echo "❌ Error: No se encontró el archivo de configuración nginx"
    echo "   Ejecuta desde el directorio raíz del proyecto: /root/PulseExpends"
    exit 1
fi

echo "📦 Creando backup de configuración actual..."
sudo mkdir -p $BACKUP_DIR
if [ -f "$NGINX_CONF_DIR/pulseexpends" ]; then
    sudo cp $NGINX_CONF_DIR/pulseexpends $BACKUP_DIR/pulseexpends.backup.$TIMESTAMP
    echo "   ✅ Backup creado: $BACKUP_DIR/pulseexpends.backup.$TIMESTAMP"
fi

# Copiar nueva configuración de Nginx
echo "📝 Copiando configuración de Nginx con paths..."
sudo cp infra/src/nginx/pulseexpends.conf $NGINX_CONF_DIR/pulseexpends

# Crear enlace simbólico si no existe
if [ ! -L "$NGINX_ENABLED_DIR/pulseexpends" ]; then
    echo "🔗 Creando enlace simbólico..."
    sudo ln -s $NGINX_CONF_DIR/pulseexpends $NGINX_ENABLED_DIR/pulseexpends
fi

# Configurar servicios systemd
echo "⚙️  Configurando servicios systemd..."

# MCP Server service
if [ -f "infra/src/scripts/pulseexpends-mcp.service" ]; then
    echo "   📦 Configurando MCP Server..."
    sudo cp infra/src/scripts/pulseexpends-mcp.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable pulseexpends-mcp.service
fi

# PDF Parser service
if [ -f "infra/src/scripts/pulseexpends-pdf-parser.service" ]; then
    echo "   📦 Configurando PDF Parser..."
    sudo cp infra/src/scripts/pulseexpends-pdf-parser.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable pulseexpends-pdf-parser.service
fi

# Status Dashboard service
if [ -f "infra/src/scripts/pulseexpends-status.service" ]; then
    echo "   📦 Configurando Status Dashboard..."
    sudo cp infra/src/scripts/pulseexpends-status.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable pulseexpends-status.service
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
        
        # Iniciar servicios
        echo "🚀 Iniciando servicios..."
        
        # MCP Server
        if systemctl is-active --quiet pulseexpends-mcp.service 2>/dev/null; then
            echo "   🔄 Reiniciando MCP Server..."
            sudo systemctl restart pulseexpends-mcp.service
        else
            echo "   ▶️  Iniciando MCP Server..."
            sudo systemctl start pulseexpends-mcp.service
        fi
        
        # PDF Parser
        if systemctl is-active --quiet pulseexpends-pdf-parser.service 2>/dev/null; then
            echo "   🔄 Reiniciando PDF Parser..."
            sudo systemctl restart pulseexpends-pdf-parser.service
        else
            echo "   ▶️  Iniciando PDF Parser..."
            sudo systemctl start pulseexpends-pdf-parser.service
        fi
        
        # Status Dashboard
        if systemctl is-active --quiet pulseexpends-status.service 2>/dev/null; then
            echo "   🔄 Reiniciando Status Dashboard..."
            sudo systemctl restart pulseexpends-status.service
        else
            echo "   ▶️  Iniciando Status Dashboard..."
            sudo systemctl start pulseexpends-status.service
        fi
        
        # Nginx
        if systemctl is-active --quiet nginx; then
            echo "   ✅ Nginx ya está activo"
        else
            echo "   ▶️  Iniciando Nginx..."
            sudo systemctl start nginx
            sudo systemctl enable nginx
        fi
        
        # Verificar estado de servicios
        echo ""
        echo "🔍 Verificando estado de servicios..."
        echo ""
        
        echo "📊 Servicio MCP Server:"
        sudo systemctl status pulseexpends-mcp.service --no-pager | grep -E "(Active|Loaded|Main PID)" || echo "   ⚠️  Servicio no encontrado"
        
        echo ""
        echo "📊 Servicio PDF Parser:"
        sudo systemctl status pulseexpends-pdf-parser.service --no-pager | grep -E "(Active|Loaded|Main PID)" || echo "   ⚠️  Servicio no encontrado"
        
        echo ""
        echo "📊 Servicio Status Dashboard:"
        sudo systemctl status pulseexpends-status.service --no-pager | grep -E "(Active|Loaded|Main PID)" || echo "   ⚠️  Servicio no encontrado"
        
        echo ""
        echo "📊 Servicio Nginx:"
        sudo systemctl status nginx --no-pager | grep -E "(Active|Loaded|Main PID)" || echo "   ⚠️  Servicio no encontrado"
        
        # Mostrar resumen
        echo ""
        echo "========================================="
        echo "✅ DESPLIEGUE COMPLETADO EXITOSAMENTE"
        echo "========================================="
        echo ""
        echo "🌐 URLs de acceso:"
        echo "   Frontend Principal: http://$DOMAIN/"
        echo "   API MCP Server:     http://$DOMAIN/mcp/"
        echo "   PDF Parser:         http://$DOMAIN/pdf/"
        echo "   Status Dashboard:   http://$DOMAIN/status/"
        echo ""
        echo "🔧 Health Checks:"
        echo "   Nginx:              http://$DOMAIN/health"
        echo "   MCP Server:         http://$DOMAIN/mcp/health"
        echo "   PDF Parser:         http://$DOMAIN/pdf/health"
        echo "   Status Dashboard:   http://$DOMAIN/status/health"
        echo ""
        echo "📋 Comandos útiles:"
        echo "   Ver logs de Nginx:      sudo journalctl -u nginx -f"
        echo "   Ver logs MCP Server:    sudo journalctl -u pulseexpends-mcp -f"
        echo "   Ver logs PDF Parser:    sudo journalctl -u pulseexpends-pdf-parser -f"
        echo "   Ver logs Status:        sudo journalctl -u pulseexpends-status -f"
        echo "   Reiniciar todos:        sudo systemctl restart nginx pulseexpends-mcp pulseexpends-pdf-parser pulseexpends-status"
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
echo "🎉 ¡Configuración completada! Los servicios están configurados con paths."
echo "   Recuerda iniciar la instancia ECS desde Huawei Cloud Console."