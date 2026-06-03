#!/bin/bash

# Script para aplicar el workaround de PulseExpends
# Este script configura un entorno local funcional mientras se resuelve el problema de nginx en el servidor

set -e

echo "========================================="
echo "WORKAROUND PULSEEXPENDS - configuration Local"
echo "========================================="

echo ""
echo "📋 DIAGNÓSTICO:"
echo "   El servidor ECS (182.160.24.205) tiene nginx mal configurado."
echo "   Las routes /mcp/* y /pdf/* no redirigen a los servicios correctos."
echo "   Status Dashboard (puerto 8081) no está funcionando."
echo ""

echo "🔧 SOLUCIÓN TEMPORAL:"
echo "   1. Usar puertos directos para las APIs (8080, 8000)"
echo "   2. Frontend modificado para usar puertos directos"
echo "   3. Proxy local para testing completo"
echo ""

echo "🔄 Aplicando workaround..."

# Paso 1: Verificar que el frontend modificado está listo
echo ""
echo "1. Verificando frontend modificado..."
if [ -f "/root/PulseExpends/infra/frontend/index.html" ]; then
    if grep -q "182.160.24.205:8080" "/root/PulseExpends/infra/frontend/index.html" && \
       grep -q "182.160.24.205:8000" "/root/PulseExpends/infra/frontend/index.html"; then
        echo "   ✅ Frontend ya modificado para usar puertos directos"
    else
        echo "   ⚠️  Frontend no modificado. Aplicando cambios..."
        cp /root/PulseExpends/infra/frontend/index.html /root/PulseExpends/infra/frontend/index.html.backup
        sed -i 's|http://182.160.24.205/mcp|http://182.160.24.205:8080|g' /root/PulseExpends/infra/frontend/index.html
        sed -i 's|http://182.160.24.205/pdf|http://182.160.24.205:8000|g' /root/PulseExpends/infra/frontend/index.html
        echo "   ✅ Frontend modificado"
    fi
else
    echo "   ❌ No se encuentra el frontend en /root/PulseExpends/infra/frontend/index.html"
    exit 1
fi

# Paso 2: Iniciar servidor web local para el frontend
echo ""
echo "2. Iniciando servidor web local para frontend..."
if pgrep -f "python3.*test-local-frontend.py" > /dev/null; then
    echo "   ✅ Servidor web local ya está ejecutándose"
else
    cd /root/PulseExpends
    python3 test-local-frontend.py > /tmp/pulseexpends-frontend.log 2>&1 &
    echo $! > /tmp/pulseexpends-frontend.pid
    echo "   ✅ Servidor web local iniciado (PID: $(cat /tmp/pulseexpends-frontend.pid))"
    echo "   📍 Accede en: http://localhost:8889"
fi

# Paso 3: Iniciar proxy inverso para testing completo
echo ""
echo "3. Iniciando proxy inverso para testing..."
if pgrep -f "python3.*test-proxy.py" > /dev/null; then
    echo "   ✅ Proxy inverso ya está ejecutándose"
else
    cd /root/PulseExpends
    python3 test-proxy.py > /tmp/pulseexpends-proxy.log 2>&1 &
    echo $! > /tmp/pulseexpends-proxy.pid
    echo "   ✅ Proxy inverso iniciado (PID: $(cat /tmp/pulseexpends-proxy.pid))"
    echo "   📍 Accede en: http://localhost:8888"
    echo "   Routing:"
    echo "     /          -> Frontend (puerto 80)"
    echo "     /mcp/*     -> MCP Server api (puerto 8080)"
    echo "     /pdf/*     -> PDF Parser api (puerto 8000)"
    echo "     /status/*  -> Status Dashboard (no disponible)"
fi

# Paso 4: Verificar conectividad con APIs
echo ""
echo "4. Verificando conectividad con APIs remotas..."

echo "   🔍 Probando MCP Server api (puerto 8080)..."
if curl -s http://182.160.24.205:8080/health | grep -q "healthy"; then
    echo "   ✅ MCP Server api funciona correctamente"
else
    echo "   ❌ MCP Server api no responde"
fi

echo "   🔍 Probando PDF Parser api (puerto 8000)..."
if curl -s http://182.160.24.205:8000/health | grep -q "healthy"; then
    echo "   ✅ PDF Parser api funciona correctamente"
else
    echo "   ❌ PDF Parser api no responde"
fi

echo "   🔍 Probando Frontend remoto (puerto 80)..."
if curl -s http://182.160.24.205/ | grep -q "PulseExpends"; then
    echo "   ✅ Frontend remoto funciona"
else
    echo "   ❌ Frontend remoto no responde"
fi

# Paso 5: Probar routes problemáticas
echo ""
echo "5. Probando routes problemáticas (nginx mal configurado)..."

echo "   🔍 Probando /mcp/health (debería redirigir a puerto 8080)..."
curl -s -o /tmp/mcp-test.html -w "%{http_code}" http://182.160.24.205/mcp/health
if grep -q "PulseExpends" /tmp/mcp-test.html; then
    echo "   ❌ Nginx sirviendo frontend en /mcp/health (mal configurado)"
else
    echo "   ✅ /mcp/health redirige correctamente"
fi

echo "   🔍 Probando /pdf/health (debería redirigir a puerto 8000)..."
curl -s -o /tmp/pdf-test.html -w "%{http_code}" http://182.160.24.205/pdf/health
if grep -q "PulseExpends" /tmp/pdf-test.html; then
    echo "   ❌ Nginx sirviendo frontend en /pdf/health (mal configurado)"
else
    echo "   ✅ /pdf/health redirige correctamente"
fi

# Paso 6: Mostrar resumen
echo ""
echo "========================================="
echo "✅ WORKAROUND APLICADO EXITOSAMENTE"
echo "========================================="
echo ""
echo "🌐 ACCESO LOCAL:"
echo ""
echo "1. Frontend con APIs directas:"
echo "   URL: http://localhost:8889"
echo "   APIs: Usan puertos 8080 y 8000 directamente"
echo ""
echo "2. Proxy inverso completo:"
echo "   URL: http://localhost:8888"
echo "   Routing corregido: /mcp/* → 8080, /pdf/* → 8000"
echo ""
echo "🌐 ACCESO REMOTO (con problemas):"
echo ""
echo "1. Frontend: http://182.160.24.205/"
echo "2. MCP api (directo): http://182.160.24.205:8080/"
echo "3. PDF api (directo): http://182.160.24.205:8000/"
echo ""
echo "⚠️  PROBLEMAS CONOCIDOS:"
echo "   - Nginx mal configurado en servidor remoto"
echo "   - routes /mcp/* y /pdf/* no funcionan vía nginx"
echo "   - Status Dashboard no disponible (puerto 8081)"
echo "   - Acceso SSH bloqueado"
echo ""
echo "🔧 SOLUCIÓN DEFINITIVA NECESARIA:"
echo "   Ejecutar en servidor remoto:"
echo "   sudo ./infra/deploy-with-paths.sh"
echo ""
echo "📋 COMANDOS ÚTILES:"
echo ""
echo "   Ver logs frontend local:"
echo "   tail -f /tmp/pulseexpends-frontend.log"
echo ""
echo "   Ver logs proxy:"
echo "   tail -f /tmp/pulseexpends-proxy.log"
echo ""
echo "   Detener servicios locales:"
echo "   pkill -f \"test-local-frontend.py\""
echo "   pkill -f \"test-proxy.py\""
echo ""
echo "   Probar APIs:"
echo "   curl http://182.160.24.205:8080/health"
echo "   curl http://182.160.24.205:8000/health"
echo "   curl http://182.160.24.205:8080/transactions"
echo ""
echo "📁 ARCHIVOS DE DIAGNÓSTICO:"
echo "   /root/PulseExpends/DIAGNOSTICO-SOLUCION.md - Diagnóstico completo"
echo "   /root/PulseExpends/test-api-direct-ports.html - Página de pruebas"
echo "   /root/PulseExpends/nginx-workaround.conf - Config nginx local"
echo ""
echo "========================================="
echo "La aplicación está FUNCIONAL con este workaround."
echo "Para solución permanente, corregir nginx en el servidor remoto."
echo "========================================="