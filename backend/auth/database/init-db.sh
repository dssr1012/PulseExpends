#!/bin/bash

# Script para inicializar la base de datos PostgreSQL para PulseExpends Auth

set -e

# Variables de configuración
DB_NAME="pulseexpends"
DB_USER="pulseexpends"
DB_PASSWORD="pulseexpends_password"
DB_HOST="localhost"
DB_PORT="5432"

echo "🔧 Inicializando base de datos para PulseExpends Auth..."

# Verificar si PostgreSQL está instalado
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL no está instalado. Instalando..."
    sudo apt-get update
    sudo apt-get install -y postgresql postgresql-contrib
fi

# Verificar si el servicio PostgreSQL está ejecutándose
if ! systemctl is-active --quiet postgresql; then
    echo "⚠️  PostgreSQL no está ejecutándose. Iniciando servicio..."
    sudo systemctl start postgresql
    sudo systemctl enable postgresql
fi

# Crear usuario y base de datos
echo "📝 Creando usuario y base de datos..."
sudo -u postgres psql <<EOF
-- Crear usuario si no existe
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
        CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
    END IF;
END
\$\$;

-- Crear base de datos si no existe
SELECT 'CREATE DATABASE $DB_NAME WITH OWNER $DB_USER ENCODING UTF8'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec

-- Conceder privilegios
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

-- Conectar a la base de datos y crear extensión uuid-ossp
\c $DB_NAME
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOF

echo "✅ Base de datos '$DB_NAME' creada con usuario '$DB_USER'"

# Ejecutar migraciones
echo "📊 Ejecutando migraciones..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f database/migrations/001_initial_schema.sql

echo "🎉 Base de datos inicializada exitosamente!"
echo ""
echo "📋 Configuración de conexión:"
echo "   Host: $DB_HOST"
echo "   Puerto: $DB_PORT"
echo "   Base de datos: $DB_NAME"
echo "   Usuario: $DB_USER"
echo "   Contraseña: $DB_PASSWORD"
echo ""
echo "🔗 URL de conexión: postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
echo ""
echo "🚀 Para ejecutar el servidor de autenticación:"
echo "   cd /root/PulseExpends/backend/auth"
echo "   go run main.go"