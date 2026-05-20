# Storage Module for PulseExpends

# Create OBS buckets
resource "huaweicloud_obs_bucket" "main" {
  for_each = var.obs_buckets
  
  bucket        = each.value.name
  storage_class = each.value.storage_class
  region        = var.region
  
  versioning {
    enabled = each.value.versioning
  }
  
  server_side_encryption {
    algorithm = "kms"
    kms_key_id = var.enable_kms ? huaweicloud_kms_key.main[0].id : null
  }
  
  dynamic "lifecycle_rule" {
    for_each = each.value.lifecycle_rules != null ? each.value.lifecycle_rules : []
    
    content {
      name    = lifecycle_rule.value.name
      prefix  = lifecycle_rule.value.prefix
      enabled = lifecycle_rule.value.enabled
      
      dynamic "expiration" {
        for_each = lifecycle_rule.value.expiration_days != null ? [1] : []
        
        content {
          days = lifecycle_rule.value.expiration_days
        }
      }
      
      dynamic "transition" {
        for_each = lifecycle_rule.value.transition_days != null ? [1] : []
        
        content {
          days          = lifecycle_rule.value.transition_days
          storage_class = lifecycle_rule.value.transition_storage_class
        }
      }
    }
  }
  
  tags = merge(var.common_tags, {
    Name        = each.value.name
    BucketType  = each.key
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create KMS key for encryption if enabled
resource "huaweicloud_kms_key" "main" {
  count = var.enable_kms ? 1 : 0
  
  key_alias       = var.kms_key_alias
  key_description = "KMS key for PulseExpends ${var.environment} environment"
  realm           = var.region
  
  pending_days = 7
  is_enabled   = true
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-kms-key"
    Purpose     = "OBS encryption"
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create database if enabled
resource "huaweicloud_rds_instance" "main" {
  count = var.enable_database ? 1 : 0
  
  name                = "${var.name_prefix}-db"
  flavor              = var.database_instance_class
  ha_replication_mode = "async"
  vpc_id              = var.vpc_id
  subnet_id           = var.private_subnet_ids[0]
  security_group_id   = var.security_group_ids[0]
  availability_zone   = [var.availability_zones[0]]
  
  db {
    type     = var.database_engine
    version  = var.database_version
    password = random_password.database[0].result
  }
  
  volume {
    type = "ULTRAHIGH"
    size = var.database_storage
  }
  
  backup_strategy {
    start_time = "03:00-04:00"
    keep_days  = var.database_backup_retention
  }
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-db"
    Engine      = var.database_engine
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create Redis instance if enabled
resource "huaweicloud_dcs_instance" "main" {
  count = var.enable_redis ? 1 : 0
  
  name               = "${var.name_prefix}-redis"
  engine             = "Redis"
  engine_version     = var.redis_engine_version
  capacity           = 2  # 2GB
  vpc_id             = var.vpc_id
  subnet_id          = var.private_subnet_ids[0]
  security_group_id  = var.security_group_ids[0]
  availability_zones = [var.availability_zones[0]]
  
  product_id = var.redis_instance_class
  
  backup_policy {
    backup_type = "auto"
    save_days   = 7
    period_type = "weekly"
    begin_at    = "02:00-03:00"
  }
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-redis"
    Engine      = "Redis"
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Generate random password for database
resource "random_password" "database" {
  count = var.enable_database ? 1 : 0
  
  length           = 16
  special          = true
  override_special = "!@#$%&*()-_=+[]{}<>:?"
}

# Store database password in OBS (secure storage)
resource "huaweicloud_obs_bucket_object" "db_password" {
  count = var.enable_database ? 1 : 0
  
  bucket       = huaweicloud_obs_bucket.main["main"].bucket
  key          = "secrets/database-password.txt"
  content      = random_password.database[0].result
  content_type = "text/plain"
  
  server_side_encryption {
    algorithm = "kms"
    kms_key_id = var.enable_kms ? huaweicloud_kms_key.main[0].id : null
  }
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-db-password"
    SecretType  = "database"
    Environment = var.environment
  })
  
  depends_on = [huaweicloud_obs_bucket.main]
}

# Outputs
output "obs_bucket_names" {
  description = "Names of the created OBS buckets"
  value       = { for k, v in huaweicloud_obs_bucket.main : k => v.bucket }
}

output "obs_bucket_urls" {
  description = "URLs of the created OBS buckets"
  value       = { for k, v in huaweicloud_obs_bucket.main : k => v.bucket_domain_name }
}

output "kms_key_id" {
  description = "ID of the KMS key"
  value       = var.enable_kms ? huaweicloud_kms_key.main[0].id : null
}

output "database_endpoint" {
  description = "Endpoint of the database"
  value       = var.enable_database ? huaweicloud_rds_instance.main[0].private_ips[0] : null
}

output "database_port" {
  description = "Port of the database"
  value       = var.enable_database ? 5432 : null
}

output "database_name" {
  description = "Database name"
  value       = var.enable_database ? "pulseexpends" : null
}

output "database_username" {
  description = "Database username"
  value       = var.enable_database ? "pulseexpends" : null
}

output "database_password_secret_location" {
  description = "Location of database password in OBS"
  value       = var.enable_database ? "obs://${huaweicloud_obs_bucket.main["main"].bucket}/secrets/database-password.txt" : null
}

output "redis_endpoint" {
  description = "Endpoint of the Redis instance"
  value       = var.enable_redis ? huaweicloud_dcs_instance.main[0].ip : null
}

output "redis_port" {
  description = "Port of the Redis instance"
  value       = var.enable_redis ? 6379 : null
}

# Data source for availability zones
data "huaweicloud_availability_zones" "available" {}