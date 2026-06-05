# ============================================================================
# Global Variables
# ============================================================================

variable "region" {
  description = "Huawei Cloud region (e.g., la-south-2 for Santiago, Chile)"
  type        = string
  default     = "la-south-2"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "pulseexpends"
}

variable "access_key" {
  description = "Huawei Cloud access key ID"
  type        = string
  sensitive   = true
}

variable "secret_key" {
  description = "Huawei Cloud secret access key"
  type        = string
  sensitive   = true
}

variable "project_id" {
  description = "Huawei Cloud project ID"
  type        = string
}

variable "agency_name" {
  description = "Huawei Cloud agency name for cross-account access (optional)"
  type        = string
  default     = ""
}

variable "agency_domain_name" {
  description = "Huawei Cloud agency domain name (optional)"
  type        = string
  default     = ""
}

# ============================================================================
# Enterprise Project Variables
# ============================================================================

variable "enable_enterprise_project" {
  description = "Whether to create and use an Enterprise Project"
  type        = bool
  default     = true
}

variable "enterprise_project_name" {
  description = "Name of the Enterprise Project to create"
  type        = string
  default     = "pulse-expendss"
}

variable "enterprise_project_type" {
  description = "Type of Enterprise Project (prod, poc, dev)"
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["prod", "poc", "dev"], var.enterprise_project_type)
    error_message = "Enterprise project type must be one of: prod, poc, dev."
  }
}

# ============================================================================
# Domain Configuration Variables
# ============================================================================

variable "enable_domain" {
  description = "Whether to configure domain name for the application"
  type        = bool
  default     = true
}

variable "domain_name" {
  description = "Domain name for the application (e.g., pulseexpends.duckdns.org)"
  type        = string
  default     = "pulseexpends.duckdns.org"
}

variable "create_dns_record" {
  description = "Whether to create DNS record (set to false if managing DNS externally)"
  type        = bool
  default     = false
}

variable "dns_zone_id" {
  description = "DNS zone ID for Huawei Cloud DNS (required if create_dns_record is true)"
  type        = string
  default     = ""
}

variable "domain_bandwidth_size" {
  description = "Maximum bandwidth size in Mbps for domain EIP (bandwidth cap when using traffic billing)"
  type        = number
  default     = 300

  validation {
    condition     = var.domain_bandwidth_size >= 1 && var.domain_bandwidth_size <= 2000
    error_message = "Domain bandwidth size must be between 1 and 2000 Mbps."
  }
}

variable "enable_ssl" {
  description = "Whether to enable SSL/TLS for the domain"
  type        = bool
  default     = false
}

variable "ssl_certificate" {
  description = "SSL certificate content (required if enable_ssl is true)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "ssl_private_key" {
  description = "SSL private key content (required if enable_ssl is true)"
  type        = string
  sensitive   = true
  default     = ""
}

# ============================================================================
# Network Variables
# ============================================================================

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnets" {
  description = "List of public subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnets" {
  description = "List of private subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24"]
}

variable "enable_nat_gateway" {
  description = "Whether to create NAT Gateway for private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Whether to use a single NAT Gateway for all private subnets"
  type        = bool
  default     = true
}

variable "enable_flow_logs" {
  description = "Whether to enable VPC Flow Logs"
  type        = bool
  default     = true
}

variable "flow_log_bucket" {
  description = "OBS bucket name for VPC Flow Logs"
  type        = string
  default     = ""
}

# ============================================================================
# Compute Variables
# ============================================================================

variable "ecs_instance_type" {
  description = "ECS instance type"
  type        = string
  default     = "c6.large.2" # 2 vCPUs, 4GB RAM

  validation {
    condition     = can(regex("^[a-z][0-9]\\.[a-z]+\\.?[0-9]*$", var.ecs_instance_type)) || can(regex("^[a-z][a-z][0-9]\\.[a-z]+\\.?[0-9]*$", var.ecs_instance_type))
    error_message = "Invalid ECS instance type format."
  }
}

variable "ecs_image_id" {
  description = "ECS image ID (Ubuntu 20.04 by default)"
  type        = string
  default     = ""
}

variable "ecs_key_pair" {
  description = "SSH key pair name for ECS instances"
  type        = string
  default     = "pulse-expends-key"
}

variable "ecs_instance_count" {
  description = "Number of ECS instances to create"
  type        = number
  default     = 1

  validation {
    condition     = var.ecs_instance_count >= 1 && var.ecs_instance_count <= 10
    error_message = "ECS instance count must be between 1 and 10."
  }
}

variable "enable_auto_scaling" {
  description = "Whether to enable auto scaling"
  type        = bool
  default     = false
}

variable "min_size" {
  description = "Minimum number of instances in auto scaling group"
  type        = number
  default     = 1
}

variable "max_size" {
  description = "Maximum number of instances in auto scaling group"
  type        = number
  default     = 3
}

variable "desired_capacity" {
  description = "Desired number of instances in auto scaling group"
  type        = number
  default     = 1
}

variable "enable_load_balancer" {
  description = "Whether to create a load balancer"
  type        = bool
  default     = false
}

variable "lb_listener_port" {
  description = "Load balancer listener port"
  type        = number
  default     = 80
}

variable "lb_protocol" {
  description = "Load balancer protocol"
  type        = string
  default     = "HTTP"

  validation {
    condition     = contains(["TCP", "HTTP", "HTTPS"], var.lb_protocol)
    error_message = "Load balancer protocol must be one of: TCP, HTTP, HTTPS."
  }
}

variable "application_port" {
  description = "Application port on ECS instances"
  type        = number
  default     = 8080
}

# ============================================================================
# Storage Variables
# ============================================================================

variable "obs_buckets" {
  description = "Map of OBS bucket configurations"
  type = map(object({
    name          = string
    storage_class = string
    versioning    = bool
    encryption    = bool
    lifecycle_rules = optional(list(object({
      name                     = string
      prefix                   = string
      enabled                  = bool
      expiration_days          = optional(number)
      transition_days          = optional(number)
      transition_storage_class = optional(string)
    })), [])
  }))

  default = {
    main = {
      name          = "pulse-expends-data"
      storage_class = "STANDARD"
      versioning    = true
      encryption    = true
      lifecycle_rules = [
        {
          name            = "temp-files"
          prefix          = "temp/"
          enabled         = true
          expiration_days = 7
        },
        {
          name                     = "logs"
          prefix                   = "logs/"
          enabled                  = true
          transition_days          = 30
          transition_storage_class = "GLACIER"
        }
      ]
    },
    documents = {
      name          = "pulse-expends-documents"
      storage_class = "STANDARD"
      versioning    = true
      encryption    = true
    }
  }
}

variable "obs_encryption_enabled" {
  description = "Whether to enable server-side encryption for OBS buckets"
  type        = bool
  default     = true
}

variable "obs_versioning_enabled" {
  description = "Whether to enable versioning for OBS buckets"
  type        = bool
  default     = true
}

variable "enable_database" {
  description = "Whether to create a database (Phase 2)"
  type        = bool
  default     = false
}

variable "database_engine" {
  description = "Database engine (postgresql, mysql, mongodb)"
  type        = string
  default     = "postgresql"

  validation {
    condition     = contains(["postgresql", "mysql", "mongodb"], var.database_engine)
    error_message = "Database engine must be one of: postgresql, mysql, mongodb."
  }
}

variable "database_version" {
  description = "Database engine version"
  type        = string
  default     = "13"
}

variable "database_instance_class" {
  description = "Database instance class"
  type        = string
  default     = "rds.pg.n1.large.2" # 2 vCPUs, 4GB RAM
}

variable "database_storage" {
  description = "Database storage size in GB"
  type        = number
  default     = 100

  validation {
    condition     = var.database_storage >= 40 && var.database_storage <= 4000
    error_message = "Database storage must be between 40 and 4000 GB."
  }
}

variable "database_backup_retention" {
  description = "Database backup retention period in days"
  type        = number
  default     = 7

  validation {
    condition     = var.database_backup_retention >= 1 && var.database_backup_retention <= 732
    error_message = "Database backup retention must be between 1 and 732 days."
  }
}

variable "enable_redis" {
  description = "Whether to create a Redis instance"
  type        = bool
  default     = false
}

variable "redis_instance_class" {
  description = "Redis instance class"
  type        = string
  default     = "redis.ha.xu1.large.r2.2" # 2 vCPUs, 4GB RAM
}

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "5.0"
}

# ============================================================================
# Monitoring Variables
# ============================================================================

variable "alert_recipients" {
  description = "List of email addresses for alert notifications"
  type        = list(string)
  default     = []
}

variable "alert_topics" {
  description = "List of SMN topic ARNs for alert notifications"
  type        = list(string)
  default     = []
}

variable "enable_dashboards" {
  description = "Whether to create Cloud Eye dashboards"
  type        = bool
  default     = true
}

# ============================================================================
# Security Variables
# ============================================================================

variable "iam_users" {
  description = "List of IAM users to create"
  type = list(object({
    name        = string
    description = optional(string)
    groups      = optional(list(string))
    policies    = optional(list(string))
  }))
  default = []
}

variable "iam_groups" {
  description = "List of IAM groups to create"
  type = list(object({
    name        = string
    description = optional(string)
    policies    = optional(list(string))
  }))
  default = []
}

variable "iam_policies" {
  description = "List of IAM policies to create"
  type = list(object({
    name        = string
    description = optional(string)
    policy      = string
  }))
  default = []
}

variable "enable_kms" {
  description = "Whether to create KMS key for encryption"
  type        = bool
  default     = true
}

variable "kms_key_alias" {
  description = "Alias for the KMS key"
  type        = string
  default     = "alias/pulse-expends"
}

variable "enable_waf" {
  description = "Whether to enable Web Application Firewall"
  type        = bool
  default     = false
}

variable "waf_policy_name" {
  description = "Name of the WAF policy"
  type        = string
  default     = "pulse-expends-waf"
}

variable "enable_antiddos" {
  description = "Whether to enable Anti-DDoS protection"
  type        = bool
  default     = true
}

# ============================================================================
# Application Variables
# ============================================================================

variable "obs_endpoint" {
  description = "OBS endpoint URL"
  type        = string
  default     = "https://obs.la-south-2.myhuaweicloud.com"
}

variable "obs_access_key" {
  description = "OBS access key (for user data script)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "obs_secret_key" {
  description = "OBS secret key (for user data script)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "obs_bucket_name" {
  description = "OBS bucket name (for user data script)"
  type        = string
  default     = "pulse-expends"
}

variable "python_service_url" {
  description = "Python PDF parser service URL"
  type        = string
  default     = "http://localhost:8000"
}

variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
  default     = ""
}

# DuckDNS Configuration

variable "duckdns_token" {
  description = "DuckDNS token for automatic DNS updates"
  type        = string
  sensitive   = true
  default     = ""
}

variable "duckdns_domain" {
  description = "DuckDNS domain name"
  type        = string
  default     = "pulseexpends.duckdns.org"
}

# ============================================================================
# RDS PostgreSQL Variables
# ============================================================================

variable "enable_rds" {
  description = "Whether to create RDS PostgreSQL database"
  type        = bool
  default     = true
}

variable "rds_instance_type" {
  description = "RDS instance type"
  type        = string
  default     = "rds.pg.n1.large.2"
}

variable "rds_storage" {
  description = "RDS storage size in GB"
  type        = number
  default     = 100

  validation {
    condition     = var.rds_storage >= 40 && var.rds_storage <= 4000
    error_message = "RDS storage must be between 40 and 4000 GB."
  }
}

variable "rds_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "17.9"

  validation {
    condition     = contains(["12", "13", "14", "15", "16", "17"], var.rds_engine_version)
    error_message = "PostgreSQL version must be one of: 12, 13, 14, 15, 16, 17, 17.9."
  }
}

variable "rds_username" {
  description = "RDS master username"
  type        = string
  default     = "pulseexpends_admin"
  sensitive   = true
}

variable "rds_password" {
  description = "RDS master password"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rds_database_name" {
  description = "RDS primary database name (auth service)"
  type        = string
  default     = "pulseexpends_auth"
}

variable "rds_core_database_name" {
  description = "RDS core database name (MCP/core service)"
  type        = string
  default     = "pulseexpends_core"
}

variable "rds_backup_retention" {
  description = "RDS backup retention period in days"
  type        = number
  default     = 7

  validation {
    condition     = var.rds_backup_retention >= 1 && var.rds_backup_retention <= 732
    error_message = "RDS backup retention must be between 1 and 732 days."
  }
}

variable "rds_backup_window" {
  description = "RDS backup window (UTC)"
  type        = string
  default     = "03:00-04:00"
}

variable "rds_maintenance_window" {
  description = "RDS maintenance window (UTC)"
  type        = string
  default     = "sun:04:00-sun:05:00"
}

variable "rds_high_availability" {
  description = "Whether to enable RDS high availability"
  type        = bool
  default     = true
}

variable "rds_ha_replication_mode" {
  description = "RDS HA replication mode"
  type        = string
  default     = "async"

  validation {
    condition     = contains(["async", "semisync", "sync"], var.rds_ha_replication_mode)
    error_message = "RDS HA replication mode must be one of: async, semisync, sync."
  }
}

variable "enable_rds_public_access" {
  description = "Whether to enable public access to RDS"
  type        = bool
  default     = false
}

variable "rds_allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access RDS"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# ============================================================================
# Custom Endpoints (for private cloud or custom regions)
# ============================================================================

variable "custom_endpoints" {
  description = "Custom endpoints for Huawei Cloud services"
  type        = map(string)
  default     = {}

  # Example:
  # custom_endpoints = {
  #   ecs = "https://ecs.example.com"
  #   vpc = "https://vpc.example.com"
  #   obs = "https://obs.example.com"
  # }
}

# ============================================================================
# Local Variables (computed)
# ============================================================================

locals {
  # Determine if this is a production environment
  is_production = var.environment == "prod"

  # Determine instance count based on environment
  actual_ecs_instance_count = local.is_production ? max(var.ecs_instance_count, 2) : var.ecs_instance_count

  # Determine if load balancer should be enabled
  actual_enable_load_balancer = local.is_production ? true : var.enable_load_balancer

  # Determine if auto scaling should be enabled
  actual_enable_auto_scaling = local.is_production ? true : var.enable_auto_scaling

  # Determine if database should be enabled
  actual_enable_database = local.is_production ? true : var.enable_database

  # Determine if Redis should be enabled
  actual_enable_redis = local.is_production ? true : var.enable_redis

  # Determine if WAF should be enabled
  actual_enable_waf = local.is_production ? true : var.enable_waf

  # Determine if RDS should be enabled
  actual_enable_rds = local.is_production ? true : var.enable_rds

  # Default image ID based on region
  default_image_id = {
    "la-south-2" = "ubuntu-20.04-server-64bit" # Ubuntu 20.04 in Santiago
    # Add more regions as needed
  }

  # Use provided image ID or default based on region
  actual_ecs_image_id = var.ecs_image_id != "" ? var.ecs_image_id : lookup(local.default_image_id, var.region, "ubuntu-20.04-server-64bit")
}