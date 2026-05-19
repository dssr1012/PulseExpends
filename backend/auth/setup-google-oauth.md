# Configuración de Google OAuth para PulseExpends

## Pasos para configurar Google OAuth

### 1. Crear un proyecto en Google Cloud Console
1. Ve a [Google Cloud Console](https://console.cloud.google.com/)
2. Crea un nuevo proyecto llamado "PulseExpends"
3. Habilita la API de Google OAuth 2.0

### 2. Configurar pantalla de consentimiento OAuth
1. En "Pantalla de consentimiento OAuth", selecciona "Externo"
2. Completa la información:
   - **Nombre de la aplicación**: PulseExpends
   - **Email de soporte**: tu-email@dominio.com
   - **Logo**: Opcional
   - **Dominios autorizados**: pulseexpends.duckdns.org
   - **URL de la política de privacidad**: http://pulseexpends.duckdns.org/privacy
   - **URL de los términos de servicio**: http://pulseexpends.duckdns.org/terms

### 3. Crear credenciales OAuth 2.0
1. Ve a "Credenciales" → "Crear credenciales" → "ID de cliente OAuth"
2. Tipo de aplicación: "Aplicación web"
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

### 4. Obtener credenciales
1. Copia el **ID de cliente** y el **Secreto de cliente**
2. Actualiza el archivo `google-oauth-config.json` con tus credenciales
3. **NO SUBAS** las credenciales reales al repositorio

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
Ejecuta el script de migración:
```bash
cd backend/auth
go run migrate.go
```

### 7. Iniciar servidor de autenticación
```bash
cd backend/auth
go run main.go
```

## Estructura de endpoints

### Backend (API)
- `POST /api/auth/register` - Registro manual
- `POST /api/auth/login` - Login manual
- `GET /api/auth/google` - Iniciar flujo OAuth
- `GET /api/auth/google/callback` - Callback OAuth
- `POST /api/auth/logout` - Cerrar sesión
- `GET /api/auth/me` - Obtener usuario actual
- `GET /api/auth/verify` - Verificar token

### Frontend
- `/login` - Página de login
- `/register` - Página de registro
- `/auth/google/callback` - Callback de Google
- `/profile` - Perfil de usuario
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

### Configuración de producción
1. Usar HTTPS en producción
2. Configurar cookies seguras (Secure, HttpOnly, SameSite)
3. Implementar rate limiting
4. Usar base de datos PostgreSQL en producción
5. Configurar backups automáticos

## Pruebas

### Local
```bash
# Iniciar servidor de autenticación
cd backend/auth
go run main.go

# Probar endpoints
curl http://localhost:8081/api/auth/health
```

### Producción
```bash
# Usar systemd para el servicio
sudo systemctl start pulseexpends-auth

# Ver logs
sudo journalctl -fu pulseexpends-auth
```

## Solución de problemas

### Error: "redirect_uri_mismatch"
Verifica que las URIs de redireccionamiento en Google Cloud Console coincidan exactamente con las configuradas.

### Error: "invalid_client"
Verifica que el ID de cliente y secreto sean correctos.

### Error: "access_denied"
Verifica que la pantalla de consentimiento esté publicada y el dominio esté autorizado.

### Error de base de datos
Verifica que PostgreSQL esté corriendo y las credenciales sean correctas.