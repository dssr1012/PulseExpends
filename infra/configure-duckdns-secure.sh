#!/bin/bash

# Script seguro para configurar DuckDNS
# NO INCLUYE EL TOKEN - debe ser proporcionado por el usuario

set -e

echo "========================================="
echo "Configuración Segura de DuckDNS"
echo "========================================="

# Variables (el token debe ser proporcionado por el usuario)
DOMAIN="pulseexpends.duckdns.org"
TOKEN=""
IP=$(curl -s https://api.ipify.org)

echo "Dominio: $DOMAIN"
echo "IP del servidor: $IP"
echo ""

# Verificar que el token esté configurado
if [ -z "$TOKEN" ]; then
    echo "❌ ERROR: Token de DuckDNS no configurado"
    echo ""
    echo "Instrucciones para obtener el token:"
    echo "1. Ve a https://www.duckdns.org"
    echo "2. Inicia sesión con tu cuenta"
    echo "3. En la página principal, copia tu token"
    echo "4. Edita este script y reemplaza TOKEN=\"\" con tu token real"
    echo ""
    echo "Ejemplo: TOKEN=\"tu_token_aqui\""
    echo ""
    echo "⚠️  IMPORTANTE: Nunca compartas tu token ni lo subas a repositorios públicos"
    exit 1
fi

echo "🔄 Actualizando dominio principal..."
RESULT=$(curl -s "https://www.duckdns.org/update?domains=$DOMAIN&token=$TOKEN&ip=$IP")
echo "Resultado: $RESULT"

echo "🔄 Actualizando subdominio api..."
RESULT=$(curl -s "https://www.duckdns.org/update?domains=api.$DOMAIN&token=$TOKEN&ip=$IP")
echo "Resultado: $RESULT"

echo "🔄 Actualizando subdominio pdf..."
RESULT=$(curl -s "https://www.duckdns.org/update?domains=pdf.$DOMAIN&token=$TOKEN&ip=$IP")
echo "Resultado: $RESULT"

echo "🔄 Actualizando subdominio status..."
RESULT=$(curl -s "https://www.duckdns.org/update?domains=status.$DOMAIN&token=$TOKEN&ip=$IP")
echo "Resultado: $RESULT"

echo ""
echo "✅ Configuración de DuckDNS completada"
echo ""
echo "🌐 URLs configuradas:"
echo "   • http://$DOMAIN"
echo "   • http://api.$DOMAIN"
echo "   • http://pdf.$DOMAIN"
echo "   • http://status.$DOMAIN"
echo ""
echo "⚠️  Notas importantes:"
echo "1. Los subdominios (api, pdf, status) deben crearse manualmente en el panel de DuckDNS"
echo "2. La propagación DNS puede tomar 5-10 minutos"
echo "3. Verifica en https://www.duckdns.org que todos los subdominios estén creados"
echo ""
echo "🔧 Para configurar actualización automática:"
echo "   ./setup-duckdns-cron.sh"
echo ""
echo "📋 Para verificar resolución DNS:"
echo "   nslookup $DOMAIN"
echo "   nslookup api.$DOMAIN"
echo "   nslookup pdf.$DOMAIN"
echo "   nslookup status.$DOMAIN"