# Huawei Cloud Credentials (DO NOT COMMIT THIS FILE)
access_key = "HPUAROTGVH59W0I1IFQL"
secret_key = "Hk7zOT6Ef6mE1nGx8Sf9bigOmL3tOzHBO8pu6v7s"
project_id = "1c42334636a749199423adad7a2d6ea3"

# Region & Environment Configuration
region       = "la-south-2"
environment  = "dev"
project_name = "pulseexpends"

# Enterprise Project Configuration
enable_enterprise_project = true
enterprise_project_id     = "9d731c5b-e130-430b-b88d-66f312596926"

# Domain Configuration
enable_domain         = true
domain_name           = "pulseexpends.duckdns.org"
create_dns_record     = false
dns_zone_id           = ""
domain_bandwidth_size = 300

# SSL Configuration
enable_ssl = false

# Network Configuration
vpc_cidr           = "10.0.0.0/16"
public_subnets     = ["10.0.1.0/24"]
private_subnets    = ["10.0.101.0/24"]
enable_nat_gateway = false
single_nat_gateway = true
enable_flow_logs   = true
flow_log_bucket    = "pulseexpends-flow-logs"

# Compute Configuration (Optimized for 10 concurrent users)
ecs_instance_type  = "s6.large.2"
ecs_instance_count = 1
ecs_disk_size      = 40
ecs_disk_type      = "SSD"
ecs_image_id       = ""
ecs_key_pair       = "pulse-expends-key"

# Database Configuration (RDS PostgreSQL) - UPDATED TO LATEST AVAILABLE 17.9
enable_rds              = true
rds_instance_type       = "rds.pg.n1.large.2"
rds_engine_version      = "17" # Updated to PostgreSQL 17.9 (confirmed available)
rds_storage             = 50
rds_backup_window       = "02:00-03:00"
rds_backup_retention    = 7
rds_database_name       = "pulseexpends"
rds_username            = "pulseexpends_admin"
rds_password            = "" # Will be generated during deployment
rds_ha_replication_mode = "async"

# Storage Configuration (OBS)
enable_obs        = true
obs_bucket_name   = "pulseexpends-documents"
obs_storage_class = "STANDARD"
obs_versioning    = true

# Load Balancer Configuration
enable_elb            = false
elb_bandwidth_size    = 5
elb_listener_protocol = "HTTP"
elb_listener_port     = 80

# Monitoring Configuration
enable_monitoring = true
monitoring_email  = ""
monitoring_phone  = ""

# Security Configuration
allowed_ssh_ips   = ["0.0.0.0/0"]
allowed_http_ips  = ["0.0.0.0/0"]
allowed_https_ips = ["0.0.0.0/0"]

# Tags
additional_tags = {
  Owner       = "DevOps"
  Department  = "Engineering"
  CostCenter  = "PulseExpends"
  Environment = "Development"
}