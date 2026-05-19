#!/bin/bash

# Script para configurar actualización automática de DuckDNS
# Requiere que configure-duckdns-secure.sh tenga el token configurado

set -e

echo "========================================="
echo "Configuración de Actualización Automática DuckDNS"
echo "========================================="

# Verificar que el script principal existe
if [ ! -f "configure-duckdns-secure.sh" ]; then
    echo "❌ ERROR: No se encontró configure-duckdns-secure.sh"
    echo "   Ejecuta desde el directorio /opt/PulseExpends/"
    exit 1
fi

# Verificar que el script principal tenga el token configurado
if grep -q 'TOKEN=""' configure-duckdns-secure.sh; then
    echo "❌ ERROR: Token no configurado en configure-duckdns-secure.sh"
    echo ""
    echo "Primero configura el token:"
    echo "1. Edita configure-duckdns-secure.sh"
    echo "2. Reemplaza TOKEN=\"\" con tu token real"
    echo "3. Guarda el archivo"
    echo ""
    echo "Ejemplo: TOKEN=\"tu_token_aqui\""
    echo ""
    echo "⚠️  IMPORTANTE: Nunca compartas tu token"
    exit 1
fi

# Crear script de actualización seguro
cat > /usr/local/bin/update-duckdns.sh << 'EOF'
#!/bin/bash
# Script seguro para actualizar DuckDNS
# Se ejecuta desde cron cada 5 minutos

DOMAIN="pulseexpends.duckdns.org"
TOKEN="TU_TOKEN_AQUI"  # Reemplazar con token real
IP=$(curl -s https://api.ipify.org 2>/dev/null || echo "0.0.0.0")

# Solo ejecutar si tenemos IP válida
if [ "$IP" != "0.0.0.0" ]; then
    curl -s "https://www.duckdns.org/update?domains=$DOMAIN,api.$DOMAIN,pdf.$DOMAIN,status.$DOMAIN&token=$TOKEN&ip=$IP" > /dev/null 2>&1
fi
EOF

# Pedir al usuario que ingrese el token
echo ""
echo "🔐 Configuración del token de DuckDNS"
echo "====================================="
echo "Por favor ingresa tu token de DuckDNS:"
read -s -p "Token: " DUCKDNS_TOKEN
echo ""

# Reemplazar el token en el script
sed -i "s/TU_TOKEN_AQUI/$DUCKDNS_TOKEN/" /usr/local/bin/update-duckdns.sh

# Hacer el script ejecutable y seguro
chmod 700 /usr/local/bin/update-duckdns.sh
chown root:root /usr/local/bin/update-duckdns.sh

echo "✅ Script de actualización creado en /usr/local/bin/update-duckdns.sh"
echo "   Permisos: 700 (solo root puede leer/ejecutar)"
echo ""

# Configurar cron job
echo "🔄 Configurando cron job..."
(crontab -l 2>/dev/null | grep -v update-duckdns.sh; echo "*/5 * * * * /usr/local/bin/update-duckdns.sh") | crontab -

echo "✅ Cron job configurado para ejecutarse cada 5 minutos"
echo ""
echo "📅 Cron jobs activos:"
crontab -l
echo ""
echo "🔍 Para ver logs de ejecución:"
echo "   grep CRON /var/log/syslog | grep update-duckdns"
echo ""
echo "🔧 Para ejecutar manualmente:"
echo "   sudo /usr/local/bin/update-duckdns.sh"
echo ""
echo "⚠️  Notas de seguridad:"
echo "1. El token está almacenado en /usr/local/bin/update-duckdns.sh"
echo "2. Solo root puede leer/ejecutar el archivo (permisos 700)"
echo "3. El cron job se ejecuta como root"
echo "4. Nunca compartas este archivo ni lo subas a repositorios públicos"
echo ""
echo "🎉 Configuración completada. DuckDNS se actualizará automáticamente cada 5 minutos."