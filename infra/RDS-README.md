# 🗄️ RDS PostgreSQL para PulseExpends

Este módulo Terraform implementa una base de datos PostgreSQL gestionada en Huawei Cloud RDS para la application PulseExpends.

## 📋 Características

### ✅ **Características Principales:**
- **PostgreSQL 15** (compatible con versiones 12-16)
- **Alta disponibilidad** con réplica asíncrona
- **Backups automáticos** con retención configurable
- **Encriptación en reposo** con KMS
- **Monitoreo integrado** con Cloud Eye
- **Escalabilidad vertical** (CPU, RAM, almacenamiento)
- **Acceso seguro** desde VPC o público (opcional)

### 🔧 **Especificaciones por Defecto:**
- **Tipo de instancia**: `rds.pg.c2.medium` (2 vCPU, 4GB RAM)
- **Almacenamiento**: 100 GB (SSD de ultra alta velocidad)
- **Versión**: PostgreSQL 15
- **Retención de backups**: 7 días
- **Ventana de backup**: 03:00-04:00 UTC
- **Ventana de mantenimiento**: Domingos 04:00-05:00 UTC

## 🚀 **implementation**

### **1. Configurar Variables**
Copia el archivo de ejemplo y personalízalo:

```bash
cp rds-example.tfvars rds.auto.tfvars
```

Edita `rds.auto.tfvars` con tus valores:

```hcl
# Credenciales de base de datos (CAMBIAR ESTOS VALORES)
rds_username = "pulseexpends_admin"
rds_password = "TuContraseñaSegura123!"  # ⚠️ CAMBIAR ESTO

# configuration de red
enable_rds_public_access = false  # Solo acceso desde VPC
rds_allowed_cidr_blocks = ["10.0.0.0/16"]  # Solo desde la VPC
```

### **2. Aplicar configuration**
```bash
# Inicializar Terraform (si no se ha hecho)
terraform init

# Ver plan de implementation
terraform plan -var-file="rds.auto.tfvars"

# Aplicar cambios
terraform apply -var-file="rds.auto.tfvars"
```

### **3. Configurar Aplicación**
Después de implementar RDS, actualiza la configuration de la application:

```bash
# En el servidor ECS
cd /opt/PulseExpends/backend/auth

# Editar .env con la conexión RDS
nano .env
```

Actualiza `DATABASE_URL`:
```bash
# Formato: postgresql://user:password@endpoint:5432/nombre_bd
DATABASE_URL=postgresql://pulseexpends_admin:TuContraseñaSegura123!@rds-endpoint:5432/pulseexpends
```

## 🔒 **Seguridad**

### **Mejores Prácticas:**
1. **Contraseñas seguras**: Usa passwords complejas de al menos 16 caracteres
2. **Acceso restringido**: Solo permite acceso desde la VPC (`enable_rds_public_access = false`)
3. **Backups automáticos**: Configura retención según necesidades (7-35 días)
4. **Monitoreo**: Habilita alertas en Cloud Eye
5. **Encriptación**: KMS está habilitado por defecto

### **configuration de Red:**
```hcl
# Solo acceso desde VPC (recomendado para producción)
enable_rds_public_access = false
rds_allowed_cidr_blocks = ["10.0.0.0/16"]

# Acceso público (solo para desarrollo/testing)
enable_rds_public_access = true
rds_allowed_cidr_blocks = ["0.0.0.0/0"]  # ⚠️ PELIGROSO para producción
```

## 💰 **Costos Estimados**

### **RDS PostgreSQL en Huawei Cloud (Chile):**
- **rds.pg.c2.medium** (2 vCPU, 4GB RAM): ~$50-70 USD/mes
- **Almacenamiento 100GB**: ~$10-15 USD/mes
- **Backups**: Incluidos en el precio
- **Transferencia de datos**: Depende del uso

### **Comparación con PostgreSQL local:**
| Característica | RDS Huawei Cloud | PostgreSQL Local |
|----------------|------------------|------------------|
| **Alta Disponibilidad** | ✅ Automática | ❌ Manual |
| **Backups** | ✅ Automáticos | ❌ Manual |
| **Parches** | ✅ Automáticos | ❌ Manual |
| **Monitoreo** | ✅ Integrado | ❌ Externo |
| **Escalabilidad** | ✅ En minutos | ❌ Horas/días |
| **Costo** | ✅ Pago por uso | ✅ Gratis (hardware) |

## 🛠️ **Operaciones**

### **Conectar a la Base de Datos:**
```bash
# Desde el servidor ECS
PGPASSWORD=TuContraseñaSegura123! psql -h rds-endpoint -U pulseexpends_admin -d pulseexpends

# Comandos útiles
\l                          # list bases de datos
\c pulseexpends            # Conectar a la base de datos
\dt                         # list tablas
\d+ tabla                  # Ver estructura de tabla
```

### **Backups y Restauración:**
- **Backups automáticos**: Diarios en ventana configurada
- **Backups manuales**: Desde consola Huawei Cloud
- **Punto en el tiempo**: Restaura a cualquier momento (hasta 7 días)
- **Exportar/Importar**: Usar `pg_dump` y `pg_restore`

### **Monitoreo:**
```bash
# Cloud Eye metrics disponibles:
- CPU utilization
- Memory utilization
- Storage usage
- Connections
- QPS (Queries per second)
- TPS (Transactions per second)
- Replication lag
- Disk I/O
```

## 📊 **Escalabilidad**

### **Escalado Vertical:**
```hcl
# En rds.auto.tfvars
rds_instance_type = "rds.pg.c2.xlarge"  # 4 vCPU, 8GB RAM
rds_storage = 200  # Aumentar almacenamiento
```

### **Escalado Horizontal:**
- **Lecturas**: Configurar réplicas de solo lectura
- **Escrituras**: update a instancia más grande
- **Particionamiento**: Implementar a nivel de application

## 🔧 **Solución de Problemas**

### **Problemas Comunes:**

#### **1. error de Conexión**
```bash
# Verificar que el security group permite el puerto 5432
# Verificar que la VPC tenga routes correctas
# Probar conexión desde ECS:
telnet rds-endpoint 5432
```

#### **2. authentication Fallida**
```bash
# Verificar credenciales
# Verificar que el user tenga permisos
# Revisar logs de RDS en Cloud Eye
```

#### **3. Alto Uso de CPU/RAM**
```bash
# Optimizar consultas
# Agregar índices
# Escalar instancia
# Revisar queries lentas
```

#### **4. Espacio en Disco Lleno**
```bash
# Limpiar datos antiguos
# Aumentar almacenamiento
# Habilitar compresión
# Archivar datos históricos
```

## 📈 **Rendimiento**

### **Parámetros Recomendados:**
```sql
-- configuration para applicationes web
ALTER SYSTEM SET shared_buffers = '1GB';
ALTER SYSTEM SET effective_cache_size = '3GB';
ALTER SYSTEM SET maintenance_work_mem = '256MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
```

### **Índices Recomendados:**
```sql
-- Para tablas de authentication
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Para tablas de transactiones
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category ON transactions(category);
```

## 🔄 **Migración desde PostgreSQL Local**

### **1. Exportar Base de Datos Local:**
```bash
pg_dump -U postgres -d pulseexpends -Fc -f backup.dump
```

### **2. create Base de Datos en RDS:**
```bash
# Conectar a RDS
PGPASSWORD=TuContraseñaSegura123! psql -h rds-endpoint -U pulseexpends_admin -d postgres -c "CREATE DATABASE pulseexpends;"
```

### **3. Importar Datos:**
```bash
pg_restore -h rds-endpoint -U pulseexpends_admin -d pulseexpends backup.dump
```

### **4. update Aplicación:**
```bash
# update DATABASE_URL en .env
DATABASE_URL=postgresql://pulseexpends_admin:TuContraseñaSegura123!@rds-endpoint:5432/pulseexpends

# Reiniciar application
sudo systemctl restart pulseexpends-auth.service
```

## 🚨 **Alertas Recomendadas**

Configurar en Cloud Eye:
- **CPU > 80%** por más de 5 minutos
- **Memoria > 90%** por más de 5 minutos
- **Espacio en disco < 20%** libre
- **Conexiones > 80%** del máximo
- **Replicación lag > 30 segundos**

## 📚 **Recursos Adicionales**

### **Documentación:**
- [Huawei Cloud RDS PostgreSQL](https://support.huaweicloud.com/intl/en-us/rds/index.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Terraform HuaweiCloud Provider](https://registry.terraform.io/providers/huaweicloud/huaweicloud/latest/docs)

### **Herramientas:**
- **pgAdmin**: GUI para administración
- **pgBadger**: Analizador de logs
- **pg_stat_statements**: Monitoreo de queries
- **pgBackRest**: Backup/restore avanzado

### **Scripts de Mantenimiento:**
```bash
# Vacuum y analyze programado
0 2 * * * PGPASSWORD=TuContraseñaSegura123! psql -h rds-endpoint -U pulseexpends_admin -d pulseexpends -c "VACUUM ANALYZE;"

# Backup diario
0 3 * * * PGPASSWORD=TuContraseñaSegura123! pg_dump -h rds-endpoint -U pulseexpends_admin -d pulseexpends -Fc > /backups/pulseexpends_$(date +\%Y\%m\%d).dump

# Rotación de backups (mantener 30 días)
0 4 * * * find /backups -name "pulseexpends_*.dump" -mtime +30 -delete
```

## 🎯 **Próximos Pasos**

1. **Implementar RDS** con `terraform apply`
2. **Configurar application** con nueva conexión
3. **Migrar datos** desde PostgreSQL local
4. **Configurar monitoreo** y alertas
5. **Establecer política de backups**
6. **Documentar procedimientos** de recuperación

## 📞 **Soporte**

### **Contactar Soporte Huawei Cloud:**
- **Portal**: https://console.huaweicloud.com/rds
- **Documentación**: https://support.huaweicloud.com/intl/en-us/rds/
- **Foros**: https://forum.huaweicloud.com/

### **Comandos de Diagnóstico:**
```bash
# Estado de RDS
terraform output rds_status
terraform output rds_endpoint

# Conexión de prueba
terraform output rds_connection_string

# Instrucciones detalladas
terraform output rds_instructions
```

---

**⚠️ IMPORTANTE**: Nunca commits credenciales en el repository. Usa variables de entorno o sistemas de gestión de secretos como Huawei Cloud KMS o Vault.