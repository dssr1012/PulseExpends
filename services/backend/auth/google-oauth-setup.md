# configuration de Google OAuth para PulseExpends

## Pasos para configurar Google OAuth

### 1. create un proyecto en Google Cloud Console
1. Ve a [Google Cloud Console](https://console.cloud.google.com/)
2. Crea un nuevo proyecto llamado "PulseExpends"
3. Habilita la api de Google OAuth 2.0

### 2. Configurar pantalla de consentimiento OAuth
1. En "Pantalla de consentimiento OAuth", selecciona "Externo"
2. Completa la information:
   - **Nombre de la application**: PulseExpends
   - **Email de soporte**: tu-email@dominio.com
   - **Logo**: Opcional
   - **Dominios autorizados**: pulseexpends.duckdns.org
   - **URL de la política de privacidad**: http://pulseexpends.duckdns.org/privacy
   - **URL de los términos de service**: http://pulseexpends.duckdns.org/terms

### 3. create credenciales OAuth 2.0
1. Ve a "Credenciales" → "create credenciales" → "ID de cliente OAuth"
2. Tipo de application: "Aplicación web"
3. Nombre: "PulseExpends Web Client"
4. URI de redireccionamiento autorizados:
   ```
   http://localhost:3000/auth/google/callback
   http://pulseexpends.duckdns.org/auth/google/callback
   http://api.pulseexpends.duckdns.org/auth/google/callback
   ```
5. Orígenes JavaScript autorizados:
   ```
   http://localhost:3000
   http://pulseexpends.duckdns.org
   https://pulseexpends.duckdns.org
   ```

### 4. get credenciales
1. Copia el **ID de cliente** y el **Secreto de cliente**
2. Actualiza el archivo `google-oauth-config.json` con tus credenciales
3. **NO SUBAS** las credenciales reales al repository

### 5. Configurar variables de entorno
Crea un archivo `.env` en el directorio `backend/`:

```bash
# Google OAuth
GOOGLE_CLIENT_ID=tu-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=tu-client-secret
GOOGLE_REDIRECT_URI=http://pulseexpends.duckdns.org/auth/google/callback

# JWT Secret (generar con: openssl rand -base64 32)
JWT_SECRET=tu-jwt-secret-aqui

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/pulseexpends
```

### 6. Configurar base de datos
Ejecuta el script de migration:
```bash
cd backend/auth
go run migrate.go
```

### 7. Iniciar servidor de authentication
```bash
cd backend/auth
go run main.go
```

## Estructura de endpoints

### Backend (api)
- `POST /api/auth/register` - Registro manual
- `POST /api/auth/login` - Login manual
- `GET /api/auth/google` - Iniciar flujo OAuth
- `GET /api/auth/google/callback` - Callback OAuth
- `POST /api/auth/logout` - Cerrar session
- `GET /api/auth/me` - get user actual
- `GET /api/auth/verify` - Verificar token

### Frontend
- `/login` - Página de login
- `/register` - Página de registro
- `/auth/google/callback` - Callback de Google
- `/profile` - Perfil de user
- `/circles` - Gestión de círculos familiares

## Seguridad

### Variables sensibles (NO SUBIR A GIT)
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET` 
- `JWT_SECRET`
- `DATABASE_URL` (si contiene credenciales)

### Archivos a ignorar
```
backend/.env
backend/auth/google-oauth-config.json
*.secret
*.key
```

### configuration de producción
1. Usar HTTPS en producción
2. Configurar cookies seguras (Secure, HttpOnly, SameSite)
3. Implementar rate limiting
4. Usar base de datos PostgreSQL en producción
5. Configurar backups automáticos

## Pruebas

### Local
```bash
# Iniciar servidor de authentication
cd backend/auth
go run main.go

# Probar endpoints
curl http://localhost:8081/api/auth/health
```

### Producción
```bash
# Usar systemd para el service
sudo systemctl start pulseexpends-auth

# Ver logs
sudo journalctl -fu pulseexpends-auth
```

## Solución de problemas

### error: "redirect_uri_mismatch"
Verifica que las URIs de redireccionamiento en Google Cloud Console coincidan exactamente con las configuradas.

### error: "invalid_client"
Verifica que el ID de cliente y secreto sean correctos.

### error: "access_denied"
Verifica que la pantalla de consentimiento esté publicada y el dominio esté autorizado.

### error de base de datos
Verifica que PostgreSQL esté corriendo y las credenciales sean correctas.