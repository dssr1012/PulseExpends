#!/bin/bash

# Script para reactivar rápidamente la infraestructura de PulseExpends
# Uso: ./reactivate-infrastructure.sh

set -e

echo "========================================="
echo "🔧 Reactivación de Infraestructura PulseExpends"
echo "========================================="
echo "Fecha: $(date)"
echo ""

# Verificar que estamos en el directorio correcto
if [ ! -f "main.tf" ]; then
    echo "❌ Error: No se encuentra main.tf"
    echo "Ejecuta desde: /root/PulseExpends-Infra/"
    exit 1
fi

# Verificar credenciales
if [ ! -f "terraform.tfvars" ] && [ ! -f "terraform.tfvars.example" ]; then
    echo "❌ Error: No se encuentra terraform.tfvars o terraform.tfvars.example"
    echo "Copia el ejemplo y configura tus credenciales:"
    echo "  cp terraform.tfvars.example terraform.tfvars"
    echo "  # Edita terraform.tfvars con tus credenciales"
    exit 1
fi

if [ ! -f "terraform.tfvars" ]; then
    echo "⚠️  Advertencia: No se encuentra terraform.tfvars"
    echo "Usando terraform.tfvars.example como base..."
    cp terraform.tfvars.example terraform.tfvars
    echo "✅ Creado terraform.tfvars desde ejemplo"
    echo "⚠️  IMPORTANTE: Edita terraform.tfvars con tus credenciales reales"
    read -p "¿Continuar? (s/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        echo "❌ Cancelado por el usuario"
        exit 0
    fi
fi

# Paso 1: Inicializar Terraform
echo ""
echo "1️⃣  Inicializando Terraform..."
terraform init

# Paso 2: Mostrar plan
echo ""
echo "2️⃣  Mostrando plan de creación..."
terraform plan

# Confirmar
echo ""
read -p "¿Crear infraestructura? (s/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Ss]$ ]]; then
    echo "❌ Cancelado por el usuario"
    exit 0
fi

# Paso 3: Aplicar configuración
echo ""
echo "3️⃣  Creando infraestructura..."
terraform apply -auto-approve

# Paso 4: Mostrar outputs
echo ""
echo "4️⃣  Mostrando outputs de Terraform..."
terraform output

# Paso 5: Obtener IP pública
echo ""
echo "5️⃣  Obteniendo IP pública del ECS..."
ECS_IP=$(terraform output -raw application_url | sed 's|http://||' | sed 's|/||')
if [ -z "$ECS_IP" ]; then
    ECS_IP=$(terraform output -raw eip_address 2>/dev/null || echo "")
fi

if [ -n "$ECS_IP" ]; then
    echo "✅ IP Pública: $ECS_IP"
    
    # Paso 6: Probar conectividad básica
    echo ""
    echo "6️⃣  Probando conectividad..."
    if ping -c 1 -W 2 "$ECS_IP" &> /dev/null; then
        echo "✅ Servidor respondiendo a ping"
        
        # Probar puerto 80
        if timeout 5 curl -s -f "http://$ECS_IP/" > /dev/null; then
            echo "✅ Servicio HTTP (puerto 80) funcionando"
            echo "🌐 URL: http://$ECS_IP/"
        else
            echo "⚠️  Servicio HTTP no disponible aún (puede tardar unos minutos)"
        fi
    else
        echo "⚠️  Servidor no responde a ping (puede estar iniciando)"
    fi
    
    # Paso 7: Instrucciones SSH
    echo ""
    echo "7️⃣  Instrucciones de conexión SSH:"
    if [ -f "pulse-expends-key.pem" ]; then
        echo "ssh -i pulse-expends-key.pem root@$ECS_IP"
    else
        echo "⚠️  No se encuentra pulse-expends-key.pem"
        echo "Usa la clave SSH configurada en Huawei Cloud"
    fi
else
    echo "⚠️  No se pudo obtener la IP pública"
fi

# Paso 8: Instrucciones para desplegar aplicación
echo ""
echo "8️⃣  Instrucciones para desplegar la aplicación:"
echo ""
echo "Una vez que el servidor esté listo:"
echo "1. Conectar al servidor:"
echo "   ssh -i pulse-expends-key.pem root@$ECS_IP"
echo ""
echo "2. Clonar/actualizar aplicación:"
echo "   cd /opt"
echo "   git clone https://github.com/dssr1012/PulseExpends.git || cd PulseExpends && git pull"
echo ""
echo "3. Ejecutar script de despliegue:"
echo "   cd /opt/PulseExpends"
echo "   ./setup-ecs-subdomains.sh"
echo ""
echo "4. Configurar DuckDNS (si es necesario):"
echo "   ./configure-duckdns-secure.sh"
echo "   # Editar /opt/PulseExpends/duckdns-update.sh con tu token"

# Paso 9: Configurar RDS (opcional)
echo ""
echo "9️⃣  Configurar RDS PostgreSQL (opcional):"
echo ""
echo "Si quieres usar RDS PostgreSQL:"
echo "1. Crear archivo de configuración:"
echo "   cp rds-example.tfvars rds.auto.tfvars"
echo "   # Editar rds.auto.tfvars con credenciales"
echo ""
echo "2. Aplicar configuración RDS:"
echo "   terraform apply -var-file=\"rds.auto.tfvars\""
echo ""
echo "3. Configurar aplicación para usar RDS:"
echo "   # En el servidor ECS:"
echo "   echo \"DATABASE_URL=postgresql://usuario:contraseña@rds-endpoint:5432/pulseexpends\" > /opt/PulseExpends/backend/auth/.env"
echo "   sudo systemctl restart pulseexpends-auth.service"

echo ""
echo "========================================="
echo "✅ Reactivación completada"
echo "========================================="
echo ""
echo "📋 Resumen:"
echo "- Infraestructura Huawei Cloud creada"
echo "- ECS Instance: ac8.large.2 (Ubuntu 20.04)"
echo "- IP Pública: $ECS_IP"
echo "- Security Groups: Configurados"
echo "- VPC: 10.0.0.0/16"
echo ""
echo "⚠️  Recordatorios:"
echo "1. Configurar DuckDNS con la nueva IP"
echo "2. Actualizar scripts de despliegue si la IP cambió"
echo "3. Verificar que todos los servicios estén corriendo"
echo ""
echo "🚀 Para probar la aplicación:"
echo "curl http://$ECS_IP/"
echo "curl http://$ECS_IP/api/auth/health"
echo "curl http://$ECS_IP/api/mcp/health"
echo ""
echo "📊 Para monitorear:"
echo "ssh -i pulse-expends-key.pem root@$ECS_IP \"systemctl status pulseexpends-* nginx\""
echo ""
echo "¡Listo! La infraestructura está activa. 🎉"