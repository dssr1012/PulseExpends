# Huawei Cloud Credentials
access_key = "HPUAELE34ORKBY58ROT4"
secret_key = "Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439"
project_id = "1c42334636a749199423adad7a2d6ea3"

# Region
region = "la-south-2"  # Santiago, Chile

# Environment
environment = "dev"  # dev, staging, prod
project_name = "pulseexpends"

# Enterprise Project Configuration
enable_enterprise_project = true
enterprise_project_name = "pulse-expendss"
enterprise_project_type = "prod"  # prod, poc, dev

# Domain Configuration
enable_domain = true
domain_name = "pulseexpends.duckdns.org"
create_dns_record = false  # Set to true if using Huawei Cloud DNS, false for external DNS like DuckDNS
dns_zone_id = ""  # Required if create_dns_record is true
domain_bandwidth_size = 300  # Mbps (pago por uso)

# SSL Configuration
enable_ssl = false  # Set to true when you have SSL certificate
# ssl_certificate = ""  # Uncomment and add SSL certificate when enable_ssl = true
# ssl_private_key = ""  # Uncomment and add SSL private key when enable_ssl = true

# Network Configuration
vpc_cidr = "10.0.0.0/16"
public_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
private_subnets = ["10.0.101.0/24", "10.0.102.0/24"]
enable_nat_gateway = true
single_nat_gateway = true
enable_flow_logs = true

# Compute Configuration (Phase 1: ECS only)
ecs_instance_type = "s6.large.2"  # 2 vCPUs, 4GB RAM - Compatible with Ubuntu
ecs_key_pair = "pulse-expends-key"
ecs_instance_count = 1
enable_load_balancer = false  # Enable for production
lb_listener_port = 80
lb_protocol = "HTTP"
application_port = 8080

# Auto Scaling (for production)
enable_auto_scaling = false
min_size = 1
max_size = 3
desired_capacity = 1

# Storage Configuration (Phase 1: OBS only)
obs_encryption_enabled = true
obs_versioning_enabled = true

# OBS Buckets Configuration
obs_buckets = {
  main = {
    name          = "pulse-expends-data-dev"
    storage_class = "STANDARD"
    versioning    = true
    encryption    = true
    lifecycle_rules = [
      {
        name    = "temp-files"
        prefix  = "temp/"
        enabled = true
        expiration_days = 7
      },
      {
        name    = "logs"
        prefix  = "logs/"
        enabled = true
        transition_days = 30
        transition_storage_class = "GLACIER"
      }
    ]
  },
  documents = {
    name          = "pulse-expends-documents-dev"
    storage_class = "STANDARD"
    versioning    = true
    encryption    = true
  }
}

# Database Configuration (Phase 2 - disabled for now)
enable_database = false
database_engine = "postgresql"
database_version = "13"
database_instance_class = "rds.pg.n1.large.2"
database_storage = 100
database_backup_retention = 7

# Redis Configuration (Phase 2 - disabled for now)
enable_redis = false
redis_instance_class = "redis.ha.xu1.large.r2.2"
redis_engine_version = "5.0"

# Monitoring Configuration
alert_recipients = ["your-email@example.com"]  # Add your email for alerts
enable_dashboards = true

# Security Configuration
enable_kms = true
kms_key_alias = "alias/pulse-expends"
enable_waf = false  # Enable for production
enable_antiddos = true

# IAM Configuration (optional)
iam_users = []
iam_groups = []
iam_policies = []

# Application Configuration
obs_endpoint = "https://obs.la-south-2.myhuaweicloud.com"
obs_access_key = "HPUAELE34ORKBY58ROT4"  # Same as access_key for OBS
obs_secret_key = "Ab4OYYfiMnhAPt8R2fdagz29y0yK5OmrCHHaO439"  # Same as secret_key for OBS
obs_bucket_name = "pulse-expends-data-dev"
python_service_url = "http://localhost:8000"
jwt_secret = "pulse-expends-jwt-secret-prod-2024-change-me"  # Secure JWT secret for authentication

# DuckDNS Configuration (for automatic DNS updates)
duckdns_token = ""  # Add your DuckDNS token here after deployment
duckdns_domain = "pulseexpends.duckdns.org"

# Custom Endpoints (for private cloud or custom regions)
custom_endpoints = {}

# Agency Configuration (optional - for cross-account access)
agency_name = ""
agency_domain_name = ""