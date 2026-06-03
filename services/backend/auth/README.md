# PulseExpends Authentication Server

Servidor de authentication y gestión de users para PulseExpends, con soporte para Google OAuth y grupos familiares (círculos).

## Características

- ✅ authentication con email/password
- ✅ authentication con Google OAuth 2.0
- ✅ Gestión de perfiles de user
- ✅ Sistema de círculos familiares
- ✅ Roles y permisos (admin, member, viewer)
- ✅ Gestión de transactiones compartidas
- ✅ api RESTful con JWT
- ✅ Base de datos PostgreSQL
- ✅ CORS configurable
- ✅ Rate limiting
- ✅ Logging estructurado

## Requisitos

- Go 1.21+
- PostgreSQL 14+
- Google OAuth credentials

## configuration

### 1. Configurar Google OAuth

1. Ve a [Google Cloud Console](https://console.cloud.google.com/)
2. Crea un nuevo proyecto o selecciona uno existente
3. Ve a "APIs & Services" → "Credentials"
4. Crea credenciales OAuth 2.0 Client ID
5. Configura las URIs de redirección autorizadas:
   - `http://localhost:8082/api/auth/google/callback`
   - `http://pulseexpends.duckdns.org/api/auth/google/callback`
   - `http://api.pulseexpends.duckdns.org/api/auth/google/callback`
6. Configura los orígenes JavaScript autorizados:
   - `http://localhost:3000`
   - `http://pulseexpends.duckdns.org`
   - `https://pulseexpends.duckdns.org`
7. Copia el Client ID y Client Secret

### 2. Configurar variables de entorno

Copia el archivo `.env.example` a `.env` y configura las variables:

```bash
cp .env.example .env
```

Edita el archivo `.env` con tus credenciales:

```env
# Google OAuth Configuration
GOOGLE_CLIENT_ID=tu-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=tu-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8082/api/auth/google/callback

# JWT Configuration
JWT_SECRET=tu-secreto-jwt-cambiar-en-produccion
JWT_EXPIRY=168h

# Database Configuration
DATABASE_URL=postgresql://pulseexpends:pulseexpends_password@localhost:5432/pulseexpends

# Server Configuration
PORT=8082
HOST=0.0.0.0
```

### 3. Inicializar base de datos

```bash
# Dar permisos de execution al script
chmod +x database/init-db.sh

# Ejecutar script de initialization
./database/init-db.sh
```

### 4. Instalar dependencias

```bash
go mod download
```

## execution

### Desarrollo

```bash
# Ejecutar con script
./run.sh

# O ejecutar directamente
go run main.go
```

### Producción con Docker

```bash
# Construir imagen
docker build -t pulseexpends-auth .

# Ejecutar contenedor
docker run -p 8082:8082 \
  -e DATABASE_URL=postgresql://pulseexpends:pulseexpends_password@host.docker.internal:5432/pulseexpends \
  -e GOOGLE_CLIENT_ID=tu-client-id \
  -e GOOGLE_CLIENT_SECRET=tu-client-secret \
  -e JWT_SECRET=tu-secreto-jwt \
  pulseexpends-auth
```

### Producción con Docker Compose

```bash
# Ejecutar todos los servicios
cd /root/PulseExpends
docker-compose up -d
```

## api Endpoints

### authentication

- `POST /api/auth/register` - Registrar nuevo user
- `POST /api/auth/login` - Iniciar session
- `GET /api/auth/google` - Iniciar session con Google
- `GET /api/auth/google/callback` - Callback de Google OAuth
- `POST /api/auth/logout` - Cerrar session
- `GET /api/auth/verify` - Verificar token JWT
- `GET /api/auth/profile` - get perfil de user
- `PUT /api/auth/profile` - update perfil
- `POST /api/auth/change-password` - Cambiar password
- `POST /api/auth/forgot-password` - Solicitar recuperación
- `POST /api/auth/reset-password` - Restablecer password

### Círculos Familiares

- `GET /api/circles` - list círculos del user
- `GET /api/circles/:id` - get círculo por ID
- `POST /api/circles` - create nuevo círculo
- `PUT /api/circles/:id` - update círculo
- `DELETE /api/circles/:id` - delete círculo
- `POST /api/circles/:id/members` - Agregar miembro
- `DELETE /api/circles/:id/members/:userId` - Remover miembro
- `PUT /api/circles/:id/members/:userId` - update rol
- `POST /api/circles/join/:code` - Unirse a círculo con código
- `POST /api/circles/:id/invite` - Invitar a círculo por email
- `GET /api/circles/:id/transactions` - Transactiones del círculo
- `GET /api/circles/:id/members` - Miembros del círculo
- `GET /api/circles/:id/activities` - Actividades del círculo

### Transactiones

- `GET /api/transactions` - list transactiones del user
- `GET /api/transactions/:id` - get transaction por ID
- `POST /api/transactions` - create nueva transaction
- `PUT /api/transactions/:id` - update transaction
- `DELETE /api/transactions/:id` - delete transaction
- `GET /api/transactions/circle/:circleId` - Transactiones de círculo
- `POST /api/transactions/:id/split` - Dividir transaction
- `POST /api/transactions/:id/approve` - Aprobar transaction
- `POST /api/transactions/:id/reject` - Rechazar transaction

### Estadísticas

- `GET /api/stats` - Estadísticas del user
- `GET /api/stats/circle/:circleId` - Estadísticas del círculo
- `GET /api/stats/monthly` - Estadísticas mensuales
- `GET /api/stats/categories` - Estadísticas por categoría

## Modelos de Datos

### user (User)
```go
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key"`
    Email        string    `gorm:"uniqueIndex;not null"`
    Username     string    `gorm:"uniqueIndex"`
    FullName     string
    AvatarURL    string
    Provider     string    `gorm:"default:'local'"`
    ProviderID   string
    IsActive     bool      `gorm:"default:true"`
    IsVerified   bool      `gorm:"default:false"`
    LastLoginAt  time.Time
    Preferences  JSONB     `gorm:"type:jsonb"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}
```

### Círculo (Circle)
```go
type Circle struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    Name        string    `gorm:"not null"`
    Description string
    Currency    string    `gorm:"default:'USD'"`
    CreatedBy   uuid.UUID `gorm:"not null"`
    IsActive    bool      `gorm:"default:true"`
    IsPublic    bool      `gorm:"default:false"`
    JoinCode    string    `gorm:"uniqueIndex"`
    Settings    JSONB     `gorm:"type:jsonb"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}
```

### Miembro de Círculo (CircleMember)
```go
type CircleMember struct {
    ID       uuid.UUID `gorm:"type:uuid;primary_key"`
    CircleID uuid.UUID `gorm:"not null"`
    UserID   uuid.UUID `gorm:"not null"`
    Role     string    `gorm:"default:'member'"`
    JoinedAt time.Time
    IsActive bool      `gorm:"default:true"`
}
```

### transaction (Transaction)
```go
type Transaction struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID       uuid.UUID `gorm:"not null"`
    CircleID     uuid.UUID
    Amount       float64   `gorm:"type:decimal(15,2);not null"`
    Currency     string    `gorm:"default:'USD'"`
    Category     string    `gorm:"not null"`
    Description  string
    Date         time.Time `gorm:"not null"`
    Type         string    `gorm:"not null"` // income, expense, transfer
    PaymentMethod string
    Location     string
    Tags         JSONB     `gorm:"type:jsonb"`
    ReceiptURL   string
    IsRecurring  bool      `gorm:"default:false"`
    RecurringID  uuid.UUID
    Status       string    `gorm:"default:'pending'"`
    ApprovedBy   uuid.UUID
    ApprovedAt   time.Time
    Notes        string
    Metadata     JSONB     `gorm:"type:jsonb"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}
```

## Seguridad

- **JWT**: Tokens firmados con clave secreta, expiración de 7 días
- **BCrypt**: Hash de passwords con costo 12
- **CORS**: Orígenes configurados explícitamente
- **Rate Limiting**: 100 solicitudes por minuto por IP
- **Cookies**: HttpOnly, Secure, SameSite=Lax
- **PostgreSQL**: Conexiones SSL/TLS en producción

## Despliegue

### 1. Configurar variables de entorno de producción

```bash
# En el servidor
export GOOGLE_CLIENT_ID=tu-client-id-produccion
export GOOGLE_CLIENT_SECRET=tu-client-secret-produccion
export JWT_SECRET=clave-secreta-fuerte-produccion
export DATABASE_URL=postgresql://user:password@servidor:5432/pulseexpends
export COOKIE_SECURE=true
export CORS_ALLOWED_ORIGINS=https://pulseexpends.duckdns.org
```

### 2. Construir y ejecutar

```bash
# Construir
go build -o auth-server main.go

# Ejecutar con systemd
sudo cp pulseexpends-auth.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable pulseexpends-auth
sudo systemctl start pulseexpends-auth
```

### 3. Configurar Nginx como reverse proxy

```nginx
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

## Monitoreo

- Health check: `GET /health`
- Métricas Prometheus: `GET /metrics`
- Logs estructurados en JSON
- Tracing de solicitudes

## Solución de Problemas

### error de conexión a PostgreSQL
```bash
# Verificar que PostgreSQL esté ejecutándose
sudo systemctl status postgresql

# Verificar credenciales
PGPASSWORD=pulseexpends_password psql -h localhost -p 5432 -U pulseexpends -d pulseexpends -c "SELECT 1"
```

### error de Google OAuth
1. Verificar que las URIs de redirección estén configuradas correctamente
2. Verificar que el Client ID y Secret sean correctos
3. Verificar que el dominio esté autorizado en Google Cloud Console

### error de JWT
```bash
# Generar nueva clave secreta
openssl rand -base64 32
```

## Licencia

MIT