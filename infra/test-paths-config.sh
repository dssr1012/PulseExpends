#!/bin/bash

# Script para probar la configuración con paths

set -e

echo "========================================="
echo "Prueba de Configuración con Paths"
echo "========================================="

DOMAIN="pulseexpends.duckdns.org"
LOCAL_IP="127.0.0.1"

echo "🔍 Probando configuración local..."
echo ""

# Verificar que Nginx esté corriendo
echo "1. Verificando Nginx..."
if systemctl is-active --quiet nginx; then
    echo "   ✅ Nginx está activo"
else
    echo "   ❌ Nginx NO está activo"
    echo "   Ejecuta: sudo systemctl start nginx"
fi

# Verificar configuración Nginx
echo ""
echo "2. Verificando configuración Nginx..."
sudo nginx -t 2>&1 | grep -q "test is successful"
if [ $? -eq 0 ]; then
    echo "   ✅ Sintaxis Nginx OK"
else
    echo "   ❌ Error en sintaxis Nginx"
    sudo nginx -t
    exit 1
fi

# Verificar servicios
echo ""
echo "3. Verificando servicios..."
SERVICES=("pulseexpends-mcp.service" "pulseexpends-pdf-parser.service" "pulseexpends-status.service")

for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet "$service" 2>/dev/null; then
        echo "   ✅ $service está activo"
    else
        echo "   ⚠️  $service NO está activo (puede ser normal si no está instalado)"
    fi
done

# Probar endpoints locales
echo ""
echo "4. Probando endpoints locales..."
echo ""

ENDPOINTS=(
    "http://$LOCAL_IP:80/health"
    "http://$LOCAL_IP:8080/health"
    "http://$LOCAL_IP:8000/health"
    "http://$LOCAL_IP:8081/"
)

for endpoint in "${ENDPOINTS[@]}"; do
    echo "   Probando: $endpoint"
    if curl -s -o /dev/null -w "%{http_code}" "$endpoint" | grep -q "200\|301\|302"; then
        echo "   ✅ OK"
    else
        echo "   ❌ Falló o no responde"
    fi
done

# Probar configuración de paths
echo ""
echo "5. Probando configuración de paths via Nginx..."
echo ""

PATHS=(
    "/"
    "/mcp/"
    "/pdf/"
    "/status/"
    "/health"
    "/mcp/health"
    "/pdf/health"
    "/status/health"
)

for path in "${PATHS[@]}"; do
    echo "   Probando: http://localhost$path"
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost$path")
    if [ "$STATUS" = "200" ] || [ "$STATUS" = "301" ] || [ "$STATUS" = "302" ]; then
        echo "   ✅ HTTP $STATUS"
    elif [ "$STATUS" = "404" ] && [ "$path" = "/" ]; then
        echo "   ⚠️  HTTP 404 (Frontend no instalado aún)"
    elif [ "$STATUS" = "000" ]; then
        echo "   ❌ No responde (servicio puede no estar corriendo)"
    else
        echo "   ⚠️  HTTP $STATUS"
    fi
done

# Verificar resolución DNS
echo ""
echo "6. Verificando DNS..."
if command -v nslookup &> /dev/null; then
    echo "   Resolviendo: $DOMAIN"
    NSLOOKUP_RESULT=$(nslookup "$DOMAIN" 2>/dev/null | grep "Address:" | tail -1 | awk '{print $2}')
    if [ -n "$NSLOOKUP_RESULT" ]; then
        echo "   ✅ Resuelve a: $NSLOOKUP_RESULT"
    else
        echo "   ⚠️  No se pudo resolver (puede ser normal si DuckDNS no está configurado)"
    fi
else
    echo "   ⚠️  nslookup no disponible"
fi

# Resumen
echo ""
echo "========================================="
echo "RESUMEN DE LA CONFIGURACIÓN"
echo "========================================="
echo ""
echo "🌐 URLs de acceso (cuando DNS esté configurado):"
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
echo "📋 Servicios configurados:"
echo "   - Nginx (puerto 80)"
echo "   - MCP Server (puerto 8080)"
echo "   - PDF Parser (puerto 8000)"
echo "   - Status Dashboard (puerto 8081)"
echo ""
echo "🚀 Para desplegar en producción:"
echo "   1. Iniciar instancia ECS en Huawei Cloud"
echo "   2. Configurar DuckDNS con IP pública"
echo "   3. Ejecutar: sudo ./infra/deploy-with-paths.sh"
echo "   4. Verificar con: curl http://$DOMAIN/health"
echo ""
echo "✅ Configuración con paths lista para usar."