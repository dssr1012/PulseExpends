

# Provider configuration - WITHOUT enterprise_project_id here (it will be set in enterprise.tf)
provider "huaweicloud" {
  region     = var.region
  access_key = var.access_key
  secret_key = var.secret_key
  project_id = var.project_id
  
  # Enterprise project configuration will be set in enterprise.tf
  # enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
  
  # Optional: Use agency for cross-account access
  agency_name = var.agency_name
  agency_domain_name = var.agency_domain_name
  
  # Optional: Specify endpoints if using custom endpoints
  endpoints = var.custom_endpoints
}

# Random resources for unique naming
resource "random_id" "deployment" {
  byte_length = 4
  prefix      = "pulse-"
}

# Local values for common configurations
locals {
  # Common tags for all resources
  common_tags = {
    Project     = "PulseExpends"
    Environment = var.environment
    Deployment  = random_id.deployment.hex
    ManagedBy   = "Terraform"
    Repository  = "https://github.com/dssr1012/PulseExpends-Infra"
  }
  
  # Naming prefix for resources
  name_prefix = "${var.project_name}-${var.environment}-${random_id.deployment.hex}"
  
  # Availability zones
  availability_zones = slice(data.huaweicloud_availability_zones.available.names, 0, min(2, length(data.huaweicloud_availability_zones.available.names)))
  
  # Enterprise project ID for modules (will be set after enterprise.tf creates it)
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# Get available availability zones
data "huaweicloud_availability_zones" "available" {}

# Network module
module "network" {
  source = "./modules/network"
  
  # Base configuration
  name_prefix     = local.name_prefix
  region          = var.region
  environment     = var.environment
  common_tags     = local.common_tags
  
  # Enterprise project configuration
  enterprise_project_id = local.enterprise_project_id
  
  # VPC configuration
  vpc_cidr        = var.vpc_cidr
  public_subnets  = var.public_subnets
  private_subnets = var.private_subnets
  
  # NAT Gateway configuration
  enable_nat_gateway = var.enable_nat_gateway
  single_nat_gateway = var.single_nat_gateway
  
  # VPC Flow Logs
  enable_flow_logs = var.enable_flow_logs
  flow_log_bucket  = var.flow_log_bucket
}

# Compute module (Phase 1: ECS)
module "compute" {
  source = "./modules/compute"
  
  # Base configuration
  name_prefix     = local.name_prefix
  region          = var.region
  environment     = var.environment
  common_tags     = local.common_tags
  
  # Enterprise project configuration
  enterprise_project_id = local.enterprise_project_id
  
  # Network configuration
  vpc_id          = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids
  security_group_ids = module.network.security_group_ids
  
  # ECS configuration
  ecs_instance_type = var.ecs_instance_type
  ecs_image_id      = var.ecs_image_id
  ecs_key_pair      = var.ecs_key_pair
  ecs_instance_count = var.ecs_instance_count
  
  # Auto Scaling configuration
  enable_auto_scaling = var.enable_auto_scaling
  min_size           = var.min_size
  max_size           = var.max_size
  desired_capacity   = var.desired_capacity
  
  # Load Balancer configuration
  enable_load_balancer = var.enable_load_balancer
  lb_listener_port    = var.lb_listener_port
  lb_protocol         = var.lb_protocol
  
  # Domain configuration
  enable_domain      = var.enable_domain
  domain_name        = var.domain_name
  domain_ip_address  = var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : null
  enable_ssl         = var.enable_ssl
  ssl_certificate_id = var.enable_ssl && var.enable_domain ? huaweicloud_waf_certificate.ssl[0].id : null
  
  # User data for application deployment
  user_data = templatefile("${path.module}/scripts/user-data.sh", {
    ENVIRONMENT      = var.environment
    REGION           = var.region
    OBS_ENDPOINT     = var.obs_endpoint
    OBS_ACCESS_KEY   = var.obs_access_key
    OBS_SECRET_KEY   = var.obs_secret_key
    OBS_BUCKET_NAME  = var.obs_bucket_name
    PYTHON_SERVICE_URL = var.python_service_url
    JWT_SECRET       = var.jwt_secret
    DOMAIN_NAME      = var.domain_name
    ENABLE_SSL       = var.enable_ssl
  })
}

# Storage module (Phase 1: OBS)
module "storage" {
  source = "./modules/storage"
  
  # Base configuration
  name_prefix     = local.name_prefix
  region          = var.region
  environment     = var.environment
  common_tags     = local.common_tags
  
  # Enterprise project configuration
  enterprise_project_id = local.enterprise_project_id
  
  # Network dependencies
  vpc_id               = module.network.vpc_id
  private_subnet_ids   = module.network.private_subnet_ids
  security_group_ids   = module.network.security_group_ids
  availability_zones   = ["la-south-2a", "la-south-2b", "la-south-2c"]
  
  # OBS configuration
  obs_buckets = var.obs_buckets
  obs_encryption_enabled = var.obs_encryption_enabled
  obs_versioning_enabled = var.obs_versioning_enabled
  
  # Database configuration (Phase 2)
  enable_database      = var.enable_database
  database_engine      = var.database_engine
  database_version     = var.database_version
  database_instance_class = var.database_instance_class
  database_storage     = var.database_storage
  database_backup_retention = var.database_backup_retention
  
  # Redis configuration
  enable_redis         = var.enable_redis
  redis_instance_class = var.redis_instance_class
  redis_engine_version = var.redis_engine_version
}

# Monitoring module
module "monitoring" {
  source = "./modules/monitoring"
  
  # Base configuration
  name_prefix     = local.name_prefix
  region          = var.region
  environment     = var.environment
  common_tags     = local.common_tags
  
  # Enterprise project configuration
  enterprise_project_id = local.enterprise_project_id
  
  # Resources to monitor
  ecs_instance_ids = module.compute.ecs_instance_ids
  load_balancer_id = module.compute.load_balancer_id
  # database_id and redis_id will be available when databases are enabled
  database_id      = null
  redis_id         = null
  
  # Alert configuration
  alert_recipients = var.alert_recipients
  alert_topics     = var.alert_topics
  
  # Dashboard configuration
  enable_dashboards = var.enable_dashboards
}

# Security module
module "security" {
  source = "./modules/security"
  
  # Base configuration
  name_prefix     = local.name_prefix
  region          = var.region
  environment     = var.environment
  common_tags     = local.common_tags
  
  # Project configuration
  project_id = var.project_id
  
  # Enterprise project configuration
  enterprise_project_id = local.enterprise_project_id
  
  # IAM configuration (disabled for now)
  iam_users       = []
  iam_groups      = []
  iam_policies    = []
  
  # KMS configuration
  enable_kms      = var.enable_kms
  kms_key_alias   = var.kms_key_alias
  
  # WAF configuration
  enable_waf      = var.enable_waf
  waf_policy_name = var.waf_policy_name
  
  # DDoS protection
  enable_antiddos = var.enable_antiddos
}

# Output values
output "deployment_id" {
  description = "Unique PulseExpends deployment ID"
  value       = random_id.deployment.hex
}

output "enterprise_project_id" {
  description = "ID of the created Enterprise Project"
  value       = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

output "enterprise_project_name" {
  description = "Name of the created Enterprise Project"
  value       = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].name : null
}

output "vpc_id" {
  description = "ID of the created VPC"
  value       = module.network.vpc_id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = module.network.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = module.network.private_subnet_ids
}

output "ecs_instance_ids" {
  description = "IDs of the ECS instances"
  value       = module.compute.ecs_instance_ids
}

output "ecs_public_ips" {
  description = "Public IP addresses of the ECS instances"
  value       = module.compute.ecs_public_ips
}

output "load_balancer_ip" {
  description = "Public IP address of the load balancer"
  value       = module.compute.load_balancer_ip
}

output "domain_ip_address" {
  description = "Public IP address for the domain"
  value       = var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : null
}

output "domain_name" {
  description = "Domain name for the application"
  value       = var.domain_name
}

output "obs_bucket_names" {
  description = "Names of the created OBS buckets"
  value       = module.storage.obs_bucket_names
}

output "database_endpoint" {
  description = "Endpoint of the database (if enabled)"
  value       = module.storage.database_endpoint
}

output "redis_endpoint" {
  description = "Endpoint of the Redis instance (if enabled)"
  value       = module.storage.redis_endpoint
}

output "monitoring_dashboard_url" {
  description = "URL of the Cloud Eye dashboard"
  value       = module.monitoring.dashboard_url
}

output "security_group_ids" {
  description = "IDs of the created security groups"
  value       = module.network.security_group_ids
}

output "kms_key_id" {
  description = "ID of the KMS key (if enabled)"
  value       = module.security.kms_key_id
}

# Application-specific outputs
output "application_url" {
  description = "URL to access the PulseExpends application"
  value       = var.enable_domain ? (var.enable_ssl ? "https://${var.domain_name}" : "http://${var.domain_name}") : (var.enable_load_balancer ? "http://${module.compute.load_balancer_ip}:${var.lb_listener_port}" : "http://${module.compute.ecs_public_ips[0]}:${var.application_port}")
}

output "application_url_with_domain" {
  description = "Full URL to access the PulseExpends application with domain"
  value       = var.enable_domain ? (var.enable_ssl ? "https://${var.domain_name}" : "http://${var.domain_name}") : null
}

output "mcp_server_url" {
  description = "URL for the MCP server"
  value       = var.enable_domain ? (var.enable_ssl ? "https://${var.domain_name}:8080" : "http://${var.domain_name}:8080") : (var.enable_load_balancer ? "http://${module.compute.load_balancer_ip}:8080" : "http://${module.compute.ecs_public_ips[0]}:8080")
}

output "python_service_url" {
  description = "URL for the Python PDF parser service"
  value       = var.enable_domain ? (var.enable_ssl ? "https://${var.domain_name}:8000" : "http://${var.domain_name}:8000") : "http://${module.compute.ecs_public_ips[0]}:8000"
}

output "ssh_access_command" {
  description = "SSH command to access the ECS instances"
  value       = length(module.compute.ecs_public_ips) > 0 ? "ssh -i ${var.ecs_key_pair}.pem root@${module.compute.ecs_public_ips[0]}" : "No ECS instances created"
}

output "dns_record_status" {
  description = "Status of DNS record creation"
  value       = var.enable_domain && var.create_dns_record && var.dns_zone_id != "" ? "DNS record created for ${var.domain_name}" : "DNS record not created (manage DNS externally or provide dns_zone_id)"
}

output "ssl_certificate_status" {
  description = "Status of SSL certificate"
  value       = var.enable_ssl && var.enable_domain && var.ssl_certificate != "" ? "SSL certificate configured" : "SSL not configured"
}

output "deployment_commands" {
  description = "Commands to deploy and manage the application"
  value = {
    deploy_application = "ssh -i ${var.ecs_key_pair}.pem root@${module.compute.ecs_public_ips[0]} 'cd /opt/pulse-expends && docker-compose up -d'"
    view_logs = "ssh -i ${var.ecs_key_pair}.pem root@${module.compute.ecs_public_ips[0]} 'cd /opt/pulse-expends && docker-compose logs -f'"
    restart_services = "ssh -i ${var.ecs_key_pair}.pem root@${module.compute.ecs_public_ips[0]} 'cd /opt/pulse-expends && docker-compose restart'"
    backup_data = "ssh -i ${var.ecs_key_pair}.pem root@${module.compute.ecs_public_ips[0]} '/opt/backup/backup.sh'"
    configure_domain = var.enable_domain ? "Configure DNS: point ${var.domain_name} to ${huaweicloud_vpc_eip.domain[0].address}" : "Domain not configured"
    setup_ssl = var.enable_ssl ? "SSL certificate configured" : "SSL not configured"
  }
}

output "next_steps" {
  description = "Next steps after deployment"
  value = {
    domain_setup = var.enable_domain ? "1. Configure DNS: Point ${var.domain_name} to ${huaweicloud_vpc_eip.domain[0].address}" : "1. Domain not configured"
    ssl_setup = var.enable_ssl ? "2. SSL certificate is configured" : "2. Configure SSL certificate if needed"
    access_application = "3. Access application at: ${var.enable_domain ? (var.enable_ssl ? "https://${var.domain_name}" : "http://${var.domain_name}") : (var.enable_load_balancer ? "http://${module.compute.load_balancer_ip}:${var.lb_listener_port}" : "http://${module.compute.ecs_public_ips[0]}:${var.application_port}")}"
    ssh_access = "4. SSH access: ssh -i ${var.ecs_key_pair}.pem root@${module.compute.ecs_public_ips[0]}"
    monitor = "5. Monitor at: ${module.monitoring.dashboard_url}"
  }
}