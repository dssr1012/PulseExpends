# Sistema de authentication PulseExpends

Este documento describe el sistema de authentication completo para PulseExpends, que incluye:

1. **Registro con email/password tradicional**
2. **Login con Google OAuth**
3. **Gestión de grupos familiares (circles) dentro de la app**
4. **authentication JWT con sessiones**
5. **Base de datos PostgreSQL para persistencia**

## 🚀 Características Principales

### 1. authentication Dual
- **Registro tradicional**: Email + password
- **Login con Google**: OAuth 2.0 con Google
- **Sesiones persistentes**: JWT tokens con refresh
- **Recuperación de password**: Email con tokens temporales

### 2. Grupos Familiares (Circles)
- **Creación de grupos**: Para compartir gastos familiares
- **Roles**: Admin, Miembro, Solo lectura
- **Códigos de invitación**: Para unirse a grupos privados
- **configuration granular**: Permisos por grupo
- **Presupuestos grupales**: Control de gastos compartidos

### 3. Seguridad
- **JWT tokens**: Firmados con secreto configurable
- **Cookies HTTP-only**: Protección contra XSS
- **CORS configurable**: Orígenes permitidos
- **Rate limiting**: Protección contra ataques de fuerza bruta
- **PostgreSQL**: Base de datos segura y escalable

## 📋 Requisitos Previos

### Servidor
- Ubuntu 20.04+ o similar
- PostgreSQL 12+
- Go 1.21+
- Nginx (para producción)

### Credenciales Google OAuth
1. create proyecto en [Google Cloud Console](https://console.cloud.google.com/)
2. Habilitar Google+ api
3. create credenciales OAuth 2.0
4. Configurar URIs de redirección

## 🛠️ Instalación Rápida

### 1. Configurar Base de Datos
```bash
# Dar permisos de execution al script
chmod +x setup-auth-database.sh

# Ejecutar script de configuration
./setup-auth-database.sh
```

### 2. Configurar Google OAuth
```bash
# Ejecutar script de configuration
./configure-google-oauth.sh

# Editar archivo .env con tus credenciales
nano backend/auth/.env
```

### 3. Iniciar Servidor de authentication
```bash
# Iniciar service
sudo systemctl start pulseexpends-auth-server

# Verificar estado
sudo systemctl status pulseexpends-auth-server

# Ver logs
sudo journalctl -u pulseexpends-auth-server -f
```

### 4. Configurar Nginx (Producción)
```nginx
# configuration para auth server
server {
    listen 80;
    server_name api.pulseexpends.duckdns.org;

    location / {
        proxy_pass http://localhost:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔧 configuration Detallada

### Variables de Entorno (.env)
```bash
# Google OAuth (REQUERIDO para login con Google)
GOOGLE_CLIENT_ID=tu-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=tu-client-secret
GOOGLE_REDIRECT_URL=http://tu-dominio.com/api/auth/google/callback

# Base de Datos
DATABASE_URL=postgresql://user:password@localhost:5432/pulseexpends

# JWT y Cookies
JWT_SECRET=generar-con-openssl-rand-hex-32
COOKIE_SECRET=generar-con-openssl-rand-hex-32

# Frontend
FRONTEND_URL=http://tu-dominio.com
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://tu-dominio.com
```

### configuration Google OAuth
1. **Google Cloud Console** → APIs & Services → Credentials
2. **Create Credentials** → OAuth 2.0 Client IDs
3. **Application type**: Web application
4. **Authorized JavaScript origins**:
   - http://localhost:3000 (desarrollo)
   - http://tu-dominio.com (producción)
5. **Authorized redirect URIs**:
   - http://localhost:8082/api/auth/google/callback (desarrollo)
   - http://tu-dominio.com/api/auth/google/callback (producción)

## 📡 Endpoints de la api

### authentication
```
POST   /api/auth/register      # Registrar user
POST   /api/auth/login         # Login con email/password
GET    /api/auth/google        # Iniciar flujo Google OAuth
GET    /api/auth/google/callback # Callback Google OAuth
POST   /api/auth/logout        # Cerrar session
GET    /api/auth/profile       # get perfil
PUT    /api/auth/profile       # update perfil
POST   /api/auth/change-password # Cambiar password
POST   /api/auth/forgot-password # Olvidé password
POST   /api/auth/reset-password  # Restablecer password
GET    /api/auth/sessions      # list sessiones activas
DELETE /api/auth/sessions/{id} # Revocar session
```

### Grupos Familiares (Circles)
```
GET    /api/circles            # list grupos del user
GET    /api/circles/{id}       # get grupo específico
POST   /api/circles            # create nuevo grupo
PUT    /api/circles/{id}       # update grupo
DELETE /api/circles/{id}       # delete grupo
GET    /api/circles/{id}/members # list miembros
POST   /api/circles/{id}/members # Agregar miembro
DELETE /api/circles/{id}/members/{userId} # Remover miembro
PUT    /api/circles/{id}/members/{userId}/role # Cambiar rol
POST   /api/circles/join/{code} # Unirse con código
POST   /api/circles/{id}/invite # Invitar por email
GET    /api/circles/{id}/transactions # Transactiones del grupo
GET    /api/circles/{id}/activities # Actividad del grupo
```

### Transactiones
```
GET    /api/transactions       # list transactiones
GET    /api/transactions/{id}  # get transaction
POST   /api/transactions       # create transaction
PUT    /api/transactions/{id}  # update transaction
DELETE /api/transactions/{id}  # delete transaction
GET    /api/transactions/circle/{circleId} # Transactiones por grupo
POST   /api/transactions/{id}/split # Dividir transaction
POST   /api/transactions/{id}/approve # Aprobar transaction
POST   /api/transactions/{id}/reject # Rechazar transaction
```

### Estadísticas
```
GET    /api/stats              # Estadísticas del user
GET    /api/stats/circle/{circleId} # Estadísticas del grupo
GET    /api/stats/monthly      # Estadísticas mensuales
GET    /api/stats/categories   # Estadísticas por categoría
```

## 🎨 Frontend - Páginas Principales

### 1. Login/Register (`/auth/login.html`)
- Formulario de registro con email/password
- Botón de login con Google
- Recuperación de password
- Credenciales de demostración

### 2. Dashboard (`/dashboard.html`)
- Resumen financiero personal
- Grupos familiares recientes
- Transactiones recientes
- Estadísticas rápidas

### 3. Grupos Familiares (`/circles.html`)
- Listado de grupos
- create nuevo grupo
- Unirse con código
- Invitar miembros
- configuration de permisos

### 4. Transactiones (`/transactions.html`)
- Listado completo de transactiones
- Filtros por categoría/tipo/date
- Agregar transactiones manuales
- Editar/delete transactiones

### 5. Subir PDF (`/upload.html`)
- Drag & drop de archivos PDF
- Extraction automática de transactiones
- Confirmación de datos extraídos

### 6. configuration (`/settings.html`)
- Preferencias de user
- Categorías personalizadas
- Alertas y notificationes
- Exportación de datos

## 🗄️ Modelos de Base de Datos

### Users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE,
    full_name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(20) DEFAULT 'local',
    provider_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### UserAuth (Contraseñas)
```sql
CREATE TABLE user_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255),
    salt VARCHAR(255),
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    last_password_change TIMESTAMP,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP
);
```

### Circles (Grupos Familiares)
```sql
CREATE TABLE circles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    currency VARCHAR(3) DEFAULT 'USD',
    created_by UUID REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    is_public BOOLEAN DEFAULT false,
    join_code VARCHAR(20) UNIQUE,
    settings JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### CircleMembers
```sql
CREATE TABLE circle_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(circle_id, user_id)
);
```

### Transactions
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    circle_id UUID REFERENCES circles(id),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    category VARCHAR(50),
    description TEXT,
    type VARCHAR(10) CHECK (type IN ('expense', 'income')),
    date TIMESTAMP NOT NULL,
    notes TEXT,
    is_approved BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## 🔒 Seguridad

### 1. JWT Tokens
- Tokens firmados con HMAC SHA256
- Expiración configurable (default: 7 días)
- Refresh tokens opcionales
- validation de firma en cada request

### 2. Cookies HTTP-only
- No accesibles desde JavaScript
- Secure flag en producción
- SameSite=Lax para protección CSRF
- Path=/ para toda la application

### 3. CORS
- Orígenes permitidos configurados
- Credenciales permitidas
- Métodos HTTP permitidos
- Headers permitidos

### 4. Rate Limiting
- Límite por IP
- Límite por user
- Ventana de tiempo configurable
- Headers de information

### 5. PostgreSQL
- Encriptación en reposo
- Backups automáticos
- Índices optimizados
- Prepared statements

## 🚨 Solución de Problemas

### error: "Database connection failed"
```bash
# Verificar PostgreSQL
sudo systemctl status postgresql

# Probar conexión
PGPASSWORD=pulseexpends_password psql -h localhost -U pulseexpends -d pulseexpends -c "SELECT 1"

# Ver logs de PostgreSQL
sudo journalctl -u postgresql -f
```

### error: "Google OAuth failed"
```bash
# Verificar credenciales
cat backend/auth/.env | grep GOOGLE

# Probar callback URL
curl -v "http://localhost:8082/api/auth/google"

# Ver logs del servidor
sudo journalctl -u pulseexpends-auth-server -f
```

### error: "CORS policy"
```bash
# Verificar CORS_ALLOWED_ORIGINS
cat backend/auth/.env | grep CORS

# Probar desde el frontend
# Abrir consola del navegador y revisar errores
```

### error: "JWT validation failed"
```bash
# Verificar JWT_SECRET
cat backend/auth/.env | grep JWT_SECRET

# Regenerar secret
openssl rand -hex 32

# update .env y reiniciar
sudo systemctl restart pulseexpends-auth-server
```

## 📈 Monitoreo

### Logs del Servidor
```bash
# Ver logs en tiempo real
sudo journalctl -u pulseexpends-auth-server -f

# Ver logs específicos
sudo journalctl -u pulseexpends-auth-server --since "1 hour ago"

# Ver logs con formato JSON
sudo journalctl -u pulseexpends-auth-server -o json
```

### Métricas de Base de Datos
```sql
-- Usuarios activos
SELECT COUNT(*) FROM users WHERE is_active = true;

-- Grupos creados
SELECT COUNT(*) FROM circles WHERE deleted_at IS NULL;

-- Transactiones por mes
SELECT DATE_TRUNC('month', created_at) as month, 
       COUNT(*) as transactions,
       SUM(amount) as total
FROM transactions 
WHERE type = 'expense'
GROUP BY month 
ORDER BY month DESC;

-- Actividad de users
SELECT u.email, COUNT(t.id) as transaction_count
FROM users u
LEFT JOIN transactions t ON u.id = t.user_id
GROUP BY u.email
ORDER BY transaction_count DESC;
```

### Health Checks
```bash
# Verificar salud del servidor
curl http://localhost:8082/api/health

# Verificar base de datos
curl http://localhost:8082/api/health?db=true

# Verificar métricas (si están habilitadas)
curl http://localhost:8082/metrics
```

## 🔄 Actualizaciones

### update Base de Datos
```bash
# Detener servidor
sudo systemctl stop pulseexpends-auth-server

# Hacer backup
pg_dump -U pulseexpends pulseexpends > backup_$(date +%Y%m%d).sql

# update código
git pull origin main

# Recompilar
cd backend/auth
go build -o auth-server main.go

# Reiniciar
sudo systemctl start pulseexpends-auth-server
```

### update Variables de Entorno
```bash
# Editar .env
nano backend/auth/.env

# Recargar configuration
sudo systemctl restart pulseexpends-auth-server
```

## 🤝 Contribución

### Reportar Issues
1. Revisar issues existentes
2. create nuevo issue con:
   - Descripción del problema
   - Pasos para reproducir
   - Logs relevantes
   - configuration del entorno

### Enviar Pull Requests
1. Fork del repository
2. create rama feature
3. Commit cambios
4. Tests y documentación
5. Pull request

### Guías de Estilo
- Go: gofmt
- JavaScript: Prettier
- SQL: UPPERCASE keywords
- Commits: Conventional commits

## 📞 Soporte

### Comunidad
- Issues en GitHub
- Discord: [enlace]
- Email: support@pulseexpends.duckdns.org

### Documentación
- [api Documentation](http://api.pulseexpends.duckdns.org/docs)
- [Frontend Guide](http://pulseexpends.duckdns.org/guide)
- [Deployment Guide](DEPLOYMENT.md)

### Contribuidores
- [@dssr1012](https://github.com/dssr1012) - Mantenedor principal

## 📄 Licencia

MIT License - Ver [LICENSE](LICENSE) para más detalles.

---

**Nota**: Este sistema está en desarrollo activo. Reporta cualquier problema o sugerencia en los issues de GitHub.