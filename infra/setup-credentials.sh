#!/bin/bash
# Script para configurar credenciales de Huawei Cloud de forma segura
# Este script crea terraform.tfvars a partir de variables de entorno

set -e  # Exit on error

echo "🔐 configuration segura de credenciales Huawei Cloud"
echo "=================================================="

# Verificar que las variables de entorno estén configuradas
if [ -z "${HUAWEI_ACCESS_KEY}" ] || [ -z "${HUAWEI_SECRET_KEY}" ] || [ -z "${HUAWEI_PROJECT_ID}" ]; then
    echo "❌ error: Variables de entorno no configuradas"
    echo ""
    echo "Por favor configura las variables de entorno:"
    echo "  export HUAWEI_ACCESS_KEY=\"tu_access_key\""
    echo "  export HUAWEI_SECRET_KEY=\"tu_secret_key\""
    echo "  export HUAWEI_PROJECT_ID=\"tu_project_id\""
    echo ""
    echo "Luego ejecuta: ./setup-credentials.sh"
    exit 1
fi

# Directorio del script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TFVARS_FILE="${SCRIPT_DIR}/terraform.tfvars"

# Verificar si ya existe terraform.tfvars
if [ -f "${TFVARS_FILE}" ]; then
    echo "⚠️  warning: ${TFVARS_FILE} ya existe"
    read -p "¿Sobrescribir? (s/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        echo "❌ cancelled por el user"
        exit 0
    fi
    # Hacer backup del archivo existente
    BACKUP_FILE="${TFVARS_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "${TFVARS_FILE}" "${BACKUP_FILE}"
    echo "📋 Backup created: ${BACKUP_FILE}"
fi

# create terraform.tfvars con las credenciales
cat > "${TFVARS_FILE}" << EOF
# Huawei Cloud Credentials
# ========================
# ARCHIVO GENERADO AUTOMÁTICAMENTE - NO COMMITEAR AL repository
# Credenciales cargadas desde variables de entorno

# Credenciales de Huawei Cloud
access_key = "${HUAWEI_ACCESS_KEY}"
secret_key = "${HUAWEI_SECRET_KEY}"
project_id = "${HUAWEI_PROJECT_ID}"

# configuration de región
region = "la-south-2"  # Santiago, Chile

# configuration de instancia ECS
instance_type = "c6.large.2"
availability_zone = "la-south-2a"
image_id = "your-image-id"  # Ubuntu 22.04 recommended
vpc_id = "your-vpc-id"
subnet_id = "your-subnet-id"
security_group_id = "your-security-group-id"
key_pair_name = "pulse-expends-key-ed25519"

# configuration de almacenamiento
system_disk_type = "SSD"
system_disk_size = 40  # GB
data_disk_type = "SSD"
data_disk_size = 100   # GB

# configuration RDS (opcional)
enable_rds = false
rds_instance_type = "rds.pg.c2.medium"
rds_storage = 100
rds_username = "pulseexpends_admin"
rds_password = "\${var.rds_password}"  # Configurar como variable de entorno
rds_availability_zone = "la-south-2a"

# configuration de red
eip_bandwidth_size = 5  # Mbps
eip_charge_mode = "traffic"

# Tags para recursos
tags = {
  Project     = "PulseExpends"
  Environment = "development"
  Owner       = "InfrastructureTeam"
  CostCenter  = "Engineering"
}

# configuration de nombres
instance_name = "pulseexpends-dev-ecs"
rds_instance_name = "pulseexpends-dev-rds"
vpc_name = "pulseexpends-vpc"
subnet_name = "pulseexpends-subnet"
security_group_name = "pulseexpends-sg"

# configuration de puertos
open_ports = [
  22,   # SSH
  80,   # HTTP
  443,  # HTTPS
  8080, # MCP Server
  8081, # Status Dashboard
  8082, # Auth api
  8083, # Main api
  8000, # PDF Parser
  5432, # PostgreSQL
]

# configuration de user
admin_username = "ubuntu"
admin_password = ""  # Usar solo SSH key

# Enterprise Project (opcional)
enterprise_project_id = ""

# Auto-scaling (opcional)
enable_auto_scaling = false
min_instance_count = 1
max_instance_count = 3

# Monitoring
enable_monitoring = true

# Backup
enable_backup = false
EOF

# Establecer permisos seguros
chmod 600 "${TFVARS_FILE}"

echo "✅ ${TFVARS_FILE} created exitosamente con permisos 600"
echo ""
echo "📋 Credenciales configuradas:"
echo "   Access Key: ${HUAWEI_ACCESS_KEY:0:8}..."
echo "   Project ID: ${HUAWEI_PROJECT_ID}"
echo "   Secret Key: ${#HUAWEI_SECRET_KEY} caracteres"
echo ""
echo "⚠️  IMPORTANTE:"
echo "   1. Este archivo NO debe committearse al repository"
echo "   2. Ya está agregado a .gitignore"
echo "   3. Para CI/CD, usa variables de entorno:"
echo "      export TF_VAR_access_key=\"\${HUAWEI_ACCESS_KEY}\""
echo "      export TF_VAR_secret_key=\"\${HUAWEI_SECRET_KEY}\""
echo "      export TF_VAR_project_id=\"\${HUAWEI_PROJECT_ID}\""
echo ""
echo "🚀 Para usar con Terraform:"
echo "   cd infra && terraform init && terraform plan"