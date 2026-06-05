# RDS PostgreSQL Configuration Example
# Copy this file to rds.auto.tfvars and customize the values

# Enable RDS PostgreSQL database
enable_rds = true

# RDS instance type
# Options: rds.pg.c2.large, rds.pg.c2.xlarge, rds.pg.c2.2xlarge, etc.
rds_instance_type = "rds.pg.c2.medium"

# Storage size in GB (40-4000 GB)
rds_storage = 100

# PostgreSQL version
# Options: "12", "13", "14", "15", "16"
rds_engine_version = "15"

# Database credentials
rds_username = "pulseexpends_admin"
rds_password = "ChangeMe123!" # CHANGE THIS TO A SECURE PASSWORD

# Database name
rds_database_name = "pulseexpends"

# Backup configuration
rds_backup_retention = 7             # Days to keep backups
rds_backup_window    = "03:00-04:00" # UTC time

# Maintenance window (UTC)
rds_maintenance_window = "sun:04:00-sun:05:00"

# High availability
rds_high_availability   = true
rds_ha_replication_mode = "async"

# Network access
enable_rds_public_access = false # Set to true if you need public access
rds_allowed_cidr_blocks = [
  "10.0.0.0/16", # VPC CIDR
  "0.0.0.0/0"    # Allow from anywhere (use with caution)
]

# Example for production:
# rds_instance_type = "rds.pg.c2.xlarge"
# rds_storage = 500
# rds_high_availability = true
# enable_rds_public_access = false
# rds_allowed_cidr_blocks = ["10.0.0.0/16"]  # Only from VPC

# Example for development:
# rds_instance_type = "rds.pg.c2.large"
# rds_storage = 100
# rds_high_availability = false
# enable_rds_public_access = true  # For easy testing
# rds_allowed_cidr_blocks = ["0.0.0.0/0"]