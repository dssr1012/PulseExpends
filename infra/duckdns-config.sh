#!/bin/bash

# Script para configurar DuckDNS con subdominios
# Este script actualiza los registros DNS en DuckDNS para todos los subdominios

set -e

echo "========================================="
echo "Configuración de DuckDNS para subdominios"
echo "========================================="

# Variables
DOMAIN="pulseexpends.duckdns.org"
TOKEN="TU_TOKEN_DE_DUCKDNS_AQUI"  # Reemplazar con tu token real
SERVER_IP="182.160.24.205"

# Verificar que el token esté configurado
if [ "$TOKEN" = "TU_TOKEN_DE_DUCKDNS_AQUI" ]; then
    echo "❌ Error: Debes configurar tu token de DuckDNS en el script"
    echo ""
    echo "Instrucciones:"
    echo "1. Ve a https://www.duckdns.org"
    echo "2. Inicia sesión con tu cuenta"
    echo "3. Copia tu token de la página principal"
    echo "4. Edita este script y reemplaza 'TU_TOKEN_DE_DUCKDNS_AQUI' con tu token real"
    echo ""
    echo "Ejemplo: TOKEN=\"abc123-def456-ghi789\""
    exit 1
fi

# Verificar conectividad
echo "🔍 Verificando conectividad con DuckDNS..."
if ! ping -c 1 www.duckdns.org > /dev/null 2>&1; then
    echo "❌ No hay conexión a internet o DuckDNS no está accesible"
    exit 1
fi

echo "✅ Conectividad verificada"

# Actualizar dominio principal
echo "🔄 Actualizando dominio principal: $DOMAIN..."
curl -s "https://www.duckdns.org/update?domains=$DOMAIN&token=$TOKEN&ip=$SERVER_IP"

if [ $? -eq 0 ]; then
    echo "✅ Dominio principal actualizado"
else
    echo "❌ Error al actualizar dominio principal"
    exit 1
fi

# Actualizar subdominios
SUBDOMAINS="api pdf status"

for SUBDOMAIN in $SUBDOMAINS; do
    FULL_DOMAIN="$SUBDOMAIN.$DOMAIN"
    echo "🔄 Actualizando subdominio: $FULL_DOMAIN..."
    curl -s "https://www.duckdns.org/update?domains=$FULL_DOMAIN&token=$TOKEN&ip=$SERVER_IP"
    
    if [ $? -eq 0 ]; then
        echo "✅ Subdominio $SUBDOMAIN actualizado"
    else
        echo "⚠️  Error al actualizar subdominio $SUBDOMAIN"
    fi
    
    # Pequeña pausa para no sobrecargar la API
    sleep 1
done

# Verificar los registros
echo ""
echo "🔍 Verificando registros DNS..."
echo "   Esto puede tomar unos minutos para propagarse"

# Esperar un momento para la propagación
sleep 5

# Verificar cada dominio
echo ""
echo "📋 Verificación de DNS:"
for DOM in $DOMAIN api.$DOMAIN pdf.$DOMAIN status.$DOMAIN; do
    echo -n "   $DOM: "
    if dig +short $DOM | grep -q "$SERVER_IP"; then
        echo "✅ OK"
    else
        echo "⏳ Pendiente de propagación"
    fi
done

# Crear archivo de configuración para cron (actualización automática)
echo ""
echo "📅 Configurando actualización automática con cron..."
CRON_JOB="*/5 * * * * curl -s 'https://www.duckdns.org/update?domains=$DOMAIN,api.$DOMAIN,pdf.$DOMAIN,status.$DOMAIN&token=$TOKEN&ip=' > /dev/null 2>&1"

# Verificar si ya existe el trabajo cron
if ! crontab -l 2>/dev/null | grep -q "duckdns.org/update"; then
    (crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -
    echo "✅ Trabajo cron agregado para actualización automática cada 5 minutos"
else
    echo "ℹ️  Trabajo cron ya existe"
fi

# Crear script de actualización manual
UPDATE_SCRIPT="/usr/local/bin/update-duckdns.sh"
sudo tee $UPDATE_SCRIPT > /dev/null << EOF
#!/bin/bash
# Script para actualizar DuckDNS manualmente
DOMAIN="$DOMAIN"
TOKEN="$TOKEN"
IP=\$(curl -s https://api.ipify.org)
curl -s "https://www.duckdns.org/update?domains=\$DOMAIN,api.\$DOMAIN,pdf.\$DOMAIN,status.\$DOMAIN&token=\$TOKEN&ip=\$IP"
echo ""
EOF

sudo chmod +x $UPDATE_SCRIPT
echo "✅ Script de actualización manual creado en $UPDATE_SCRIPT"

echo ""
echo "========================================="
echo "✅ CONFIGURACIÓN DE DUCKDNS COMPLETADA"
echo "========================================="
echo ""
echo "🌐 Dominios configurados:"
echo "   - $DOMAIN"
echo "   - api.$DOMAIN"
echo "   - pdf.$DOMAIN"
echo "   - status.$DOMAIN"
echo ""
echo "📡 IP del servidor: $SERVER_IP"
echo ""
echo "⚙️  Configuración automática:"
echo "   - DuckDNS se actualizará automáticamente cada 5 minutos"
echo "   - Para actualizar manualmente: sudo $UPDATE_SCRIPT"
echo ""
echo "⚠️  NOTA IMPORTANTE:"
echo "   La propagación DNS puede tomar hasta 5-10 minutos."
echo "   Si la instancia ECS está apagada, los servicios no estarán disponibles."
echo ""
echo "🔧 Próximos pasos:"
echo "   1. Inicia la instancia ECS desde Huawei Cloud Console"
echo "   2. Ejecuta el script de despliegue: ./deploy-subdomains.sh"
echo "   3. Espera la propagación DNS (5-10 minutos)"
echo "   4. Accede a http://$DOMAIN para verificar"