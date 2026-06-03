# 🔐 Gestión Segura de Credenciales - PulseExpends

Este documento describe cómo gestionar credenciales de forma segura en el proyecto PulseExpends.

## 📋 Archivos Sensibles Excluidos

Los siguientes archivos están excluidos del repositorio (ver `.gitignore`):

```
# Claves y certificados
*.pem
*.key
pulse-expends-key
pulse-expends-key.pub

# Credenciales de Huawei Cloud
terraform.tfvars
*.tfvars

# Variables de entorno
.env
.env.local
secrets.*
credentials.*

# Tokens y secrets
*.token
*.secret
access_keys.txt

# Directorios de claves
private_keys/
ssh-keys/

# Archivos de configuración local
config.local.*
settings.local.*

# Backups de credenciales
*.backup.*
credentials-backup.*
```

## 🚀 Configuración Inicial Segura

### 1. Generar Nueva Clave SSH
```bash
cd infra
ssh-keygen -t ed25519 -f pulse-expends-key -N "" -C "pulse-expends-ecs-key-$(date +%Y%m%d)"
chmod 600 pulse-expends-key
```

### 2. Configurar Credenciales Huawei Cloud
```bash
# Opción A: Variables de entorno (RECOMENDADO)
export HUAWEI_ACCESS_KEY="tu_access_key"
export HUAWEI_SECRET_KEY="tu_secret_key"
export HUAWEI_PROJECT_ID="tu_project_id"

# Opción B: Usar script de configuración
cd infra
./setup-credentials.sh
```

### 3. Importar Clave SSH a Huawei Cloud
1. Ir a **Huawei Cloud Console** → **ECS** → **Key Pairs**
2. Click en **"Import Key Pair"**
3. Nombre: `pulse-expends-key-ed25519`
4. Pegar contenido de `infra/pulse-expends-key.pub`
5. Click **"OK"**

## 🔧 Script de Configuración

El script `infra/setup-credentials.sh` crea `terraform.tfvars` de forma segura:

```bash
# Configurar variables de entorno primero
export HUAWEI_ACCESS_KEY="tu_access_key"
export HUAWEI_SECRET_KEY="tu_secret_key"
export HUAWEI_PROJECT_ID="tu_project_id"

# Ejecutar script
cd infra
chmod +x setup-credentials.sh
./setup-credentials.sh
```

## 📁 Estructura de Archivos

```
infra/
├── terraform.tfvars.example    # Template con placeholders
├── setup-credentials.sh        # Script para generar terraform.tfvars
├── terraform.tfvars            # GENERADO AUTOMÁTICAMENTE (NO COMMIT)
├── pulse-expends-key           # Clave SSH privada (NO COMMIT)
└── pulse-expends-key.pub       # Clave SSH pública (puede committearse)
```

## 🔒 Buenas Prácticas

### Nunca Committear:
- ✅ `terraform.tfvars` (usar `terraform.tfvars.example`)
- ✅ `.env` files (usar `.env.example`)
- ✅ Claves privadas (`.pem`, `.key`)
- ✅ Tokens y secrets

### Siempre Committear:
- ✅ `terraform.tfvars.example` (con placeholders)
- ✅ `.env.example` (con valores de ejemplo)
- ✅ Claves públicas (`.pub`)
- ✅ Scripts de configuración

### Para CI/CD:
```bash
# Usar variables de entorno
export TF_VAR_access_key="${HUAWEI_ACCESS_KEY}"
export TF_VAR_secret_key="${HUAWEI_SECRET_KEY}"
export TF_VAR_project_id="${HUAWEI_PROJECT_ID}"

# O usar backend remoto para state
terraform {
  backend "s3" {
    bucket = "tu-bucket"
    key    = "terraform.tfstate"
    region = "la-south-2"
  }
}
```

## 🚨 Rotación de Credenciales

Si sospechas que las credenciales están comprometidas:

### 1. Huawei Cloud:
1. Ir a **IAM** → **Users** → Tu usuario
2. **Security credentials** → **Access Keys**
3. Revocar la key comprometida
4. Generar nueva Access Key
5. Actualizar `terraform.tfvars`

### 2. Clave SSH:
```bash
# Generar nueva clave
ssh-keygen -t ed25519 -f pulse-expends-key-new -N ""

# Importar a Huawei Cloud
# Bind a la instancia ECS

# Eliminar clave antigua
rm pulse-expends-key pulse-expends-key.pub
```

### 3. Otras Credenciales:
- **Google OAuth**: Google Cloud Console → APIs & Services → Credentials
- **JWT Secrets**: Regenerar con `openssl rand -base64 32`
- **Database**: Rotar contraseñas en RDS

## 📞 Soporte

### Problemas Comunes:

**Error: "Access denied" al conectar por SSH**
```bash
# Verificar que la clave está bindeada a la instancia
# En Huawei Cloud Console: ECS → Instances → Tu instancia → More → Change Key Pair
```

**Error: "Invalid credentials" en Terraform**
```bash
# Verificar variables de entorno
echo $HUAWEI_ACCESS_KEY
echo $HUAWEI_PROJECT_ID

# O verificar terraform.tfvars
cat infra/terraform.tfvars | head -3
```

**Error: "Permission denied" en archivos**
```bash
# Establecer permisos correctos
chmod 600 infra/terraform.tfvars
chmod 600 infra/pulse-expends-key
```

## 🔍 Verificación de Seguridad

Para verificar que no hay credenciales expuestas:
```bash
# Buscar posibles leaks
cd /root/PulseExpends
grep -r "password\|secret\|key\|token" \
  --include="*.go" --include="*.py" --include="*.js" \
  --include="*.ts" --include="*.json" --include="*.yaml" \
  --include="*.yml" --include="*.tf" . 2>/dev/null | \
  grep -v node_modules | \
  grep -v ".git" | \
  grep -v "example" | \
  grep -v "test" | \
  grep -v "placeholder" | \
  grep -v "change-this" | \
  grep -v "your-"
```

## 🎯 Resumen

1. **Usar variables de entorno** para credenciales
2. **Nunca committear** `terraform.tfvars` o `.env`
3. **Usar templates** con placeholders (`.example`)
4. **Rotar credenciales** periódicamente
5. **Revisar `.gitignore`** regularmente

**Mantén las credenciales seguras!** 🔐