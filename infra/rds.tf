# RDS PostgreSQL Database for PulseExpends Application
resource "huaweicloud_rds_instance" "postgresql" {
  count = var.enable_rds ? 1 : 0

  name              = "${local.name_prefix}-rds"
  flavor            = var.rds_instance_type
  vpc_id            = huaweicloud_vpc.main.id
  subnet_id         = huaweicloud_vpc_subnet.private.id # Use private subnet
  security_group_id = huaweicloud_networking_secgroup.main.id

  availability_zone = ["la-south-2a"]

  db {
    type     = "PostgreSQL"
    version  = var.rds_engine_version
    password = var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result
  }

  volume {
    type = "ESSD"
    size = var.rds_storage
  }

  backup_strategy {
    start_time = var.rds_backup_window
    keep_days  = var.rds_backup_retention
  }

  tags = merge(local.common_tags, {
    Name    = "${local.name_prefix}-rds"
    Purpose = "Application Database"
    Engine  = "PostgreSQL"
  })

  enterprise_project_id = "9d731c5b-e130-430b-b88d-66f312596926"

  lifecycle {
    ignore_changes = [
      db[0].password # Password changes should be handled separately
    ]
  }
}

# Generate random password for RDS if not provided
resource "random_password" "rds_password" {
  count = var.enable_rds && var.rds_password == "" ? 1 : 0

  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# Database for PulseExpends Auth service
resource "huaweicloud_rds_database" "pulseexpends_auth" {
  count = var.enable_rds ? 1 : 0

  instance_id   = huaweicloud_rds_instance.postgresql[0].id
  name          = var.rds_database_name
  character_set = "UTF8"

  depends_on = [huaweicloud_rds_instance.postgresql]
}

# Database for PulseExpends Core/MCP service
resource "huaweicloud_rds_database" "pulseexpends_core" {
  count = var.enable_rds ? 1 : 0

  instance_id   = huaweicloud_rds_instance.postgresql[0].id
  name          = var.rds_core_database_name
  character_set = "UTF8"

  depends_on = [huaweicloud_rds_instance.postgresql]
}

# Database user with full privileges
resource "huaweicloud_rds_account" "app_user" {
  count = var.enable_rds ? 1 : 0

  instance_id = huaweicloud_rds_instance.postgresql[0].id
  name        = var.rds_username
  password    = var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result

  depends_on = [huaweicloud_rds_instance.postgresql]
}

# Security group rule for PostgreSQL access - ONLY from VPC CIDR
resource "huaweicloud_networking_secgroup_rule" "postgresql" {
  count = var.enable_rds ? 1 : 0

  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 5432
  port_range_max    = 5432
  remote_ip_prefix  = var.vpc_cidr # Restrict to VPC CIDR only
  security_group_id = huaweicloud_networking_secgroup.main.id

  description = "Allow PostgreSQL access from within VPC"
}

# Output RDS connection details
output "rds_instance_id" {
  description = "ID of the RDS instance"
  value       = var.enable_rds ? huaweicloud_rds_instance.postgresql[0].id : null
}

output "rds_endpoint" {
  description = "Endpoint of the RDS instance"
  value       = var.enable_rds ? huaweicloud_rds_instance.postgresql[0].private_ips[0] : null
}

output "rds_port" {
  description = "Port of the RDS instance"
  value       = var.enable_rds ? 5432 : null
}

output "rds_database_name" {
  description = "Name of the RDS database"
  value       = var.enable_rds ? var.rds_database_name : null
}

output "rds_username" {
  description = "Username for the RDS database"
  value       = var.enable_rds ? var.rds_username : null
  sensitive   = true
}

output "rds_password" {
  description = "Password for the RDS database (auto-generated if not provided)"
  value       = var.enable_rds ? (var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result) : null
  sensitive   = true
}

output "rds_core_database_name" {
  description = "Name of the RDS core database (MCP service)"
  value       = var.enable_rds ? var.rds_core_database_name : null
}

output "rds_connection_string" {
  description = "PostgreSQL connection string for auth database"
  value       = var.enable_rds ? "postgresql://${var.rds_username}:${var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result}@${huaweicloud_rds_instance.postgresql[0].private_ips[0]}:5432/${var.rds_database_name}" : null
  sensitive   = true
}

output "rds_core_connection_string" {
  description = "PostgreSQL connection string for core database"
  value       = var.enable_rds ? "postgresql://${var.rds_username}:${var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result}@${huaweicloud_rds_instance.postgresql[0].private_ips[0]}:5432/${var.rds_core_database_name}" : null
  sensitive   = true
}

output "rds_public_connection_string" {
  description = "PostgreSQL public connection string for auth database (if enabled)"
  value       = var.enable_rds && var.enable_rds_public_access ? "postgresql://${var.rds_username}:${var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result}@${huaweicloud_rds_instance.postgresql[0].public_ips[0]}:5432/${var.rds_database_name}" : null
  sensitive   = true
}

output "rds_status" {
  description = "Status of RDS deployment"
  value       = var.enable_rds ? "ENABLED" : "DISABLED"
}

output "rds_instructions" {
  description = "Instructions for connecting to RDS"
  value       = var.enable_rds ? "\n📋 RDS PostgreSQL Configuration:\n===============================\nInstance: ${huaweicloud_rds_instance.postgresql[0].name}\nEndpoint: ${huaweicloud_rds_instance.postgresql[0].private_ips[0]}:5432\n\nAuth Database: ${var.rds_database_name}\nCore Database: ${var.rds_core_database_name}\nUsername: ${var.rds_username}\nPassword: ${var.rds_password != "" ? "[provided]" : "[auto-generated]"}\n\nAuth Connection String:\npostgresql://${var.rds_username}:***@${huaweicloud_rds_instance.postgresql[0].private_ips[0]}:5432/${var.rds_database_name}\n\nCore Connection String:\npostgresql://${var.rds_username}:***@${huaweicloud_rds_instance.postgresql[0].private_ips[0]}:5432/${var.rds_core_database_name}\n\nTo connect from ECS instance:\nPGPASSWORD=${var.rds_password != "" ? var.rds_password : random_password.rds_password[0].result} psql -h ${huaweicloud_rds_instance.postgresql[0].private_ips[0]} -U ${var.rds_username} -d ${var.rds_database_name}\n" : "RDS is disabled. Set enable_rds = true to create a PostgreSQL database."
  sensitive   = true
}
