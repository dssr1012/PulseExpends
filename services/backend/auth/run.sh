#!/bin/bash

# Script para ejecutar el servidor de autenticación PulseExpends

set -e

echo "🚀 Iniciando servidor de autenticación PulseExpends..."

# Verificar que las variables de entorno estén configuradas
if [ ! -f .env ]; then
    echo "⚠️  Archivo .env no encontrado. Creando desde .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ Archivo .env creado desde .env.example"
        echo "⚠️  Por favor, configura las variables de entorno en .env antes de continuar"
        exit 1
    else
        echo "❌ Archivo .env.example no encontrado"
        exit 1
    fi
fi

# Cargar variables de entorno
export $(grep -v '^#' .env | xargs)

# Verificar que PostgreSQL esté ejecutándose
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "⚠️  PostgreSQL no está ejecutándose. Iniciando..."
    sudo systemctl start postgresql
    sleep 2
fi

# Verificar que la base de datos exista
if ! PGPASSWORD=$DATABASE_PASSWORD psql -h localhost -p 5432 -U pulseexpends -d pulseexpends -c "SELECT 1" > /dev/null 2>&1; then
    echo "⚠️  Base de datos no encontrada. Ejecutando migraciones..."
    ./database/init-db.sh
fi

# Instalar dependencias si es necesario
echo "📦 Verificando dependencias..."
go mod download

# Construir el servidor
echo "🔨 Construyendo servidor..."
go build -o auth-server main.go

# Ejecutar el servidor
echo "🌐 Iniciando servidor en http://localhost:$PORT"
echo "📊 Health check: http://localhost:$PORT/health"
echo "📚 API Docs: http://localhost:$PORT/swagger/index.html"
echo ""
echo "📋 Configuración:"
echo "   Port: $PORT"
echo "   Database: $DATABASE_URL"
echo "   Google OAuth: $ENABLE_GOOGLE_OAUTH"
echo "   CORS Origins: $CORS_ALLOWED_ORIGINS"
echo ""
echo "🛑 Presiona Ctrl+C para detener el servidor"

# Ejecutar el servidor
./auth-server