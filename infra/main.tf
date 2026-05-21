# Minimal Terraform configuration for PulseExpends
# This creates only the essential resources to get started

terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    huaweicloud = {
      source  = "huaweicloud/huaweicloud"
      version = "~> 1.91"
    }
    
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.0"
    }
  }
  
  backend "local" {
    path = "terraform.tfstate"
  }
}

# Provider configuration
provider "huaweicloud" {
  region     = var.region
  access_key = var.access_key
  secret_key = var.secret_key
  project_id = var.project_id
  
  # Optional: Use agency for cross-account access
  agency_name = var.agency_name
  agency_domain_name = var.agency_domain_name
}

# Random ID for unique naming
resource "random_id" "deployment" {
  byte_length = 4
  prefix      = "pulse-"
}

# Local values for common configuration
locals {
  name_prefix = "${var.project_name}-${var.environment}-${random_id.deployment.hex}"
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
    Deployment  = random_id.deployment.hex
  }
}

# Enterprise Project
resource "huaweicloud_enterprise_project" "pulse_expends" {
  count = var.enable_enterprise_project ? 1 : 0
  
  name        = var.enterprise_project_name
  description = "Enterprise Project for PulseExpends application"
  type        = var.enterprise_project_type
}

# VPC
resource "huaweicloud_vpc" "main" {
  name = "${local.name_prefix}-vpc"
  cidr = var.vpc_cidr
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-vpc"
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# Public Subnet
resource "huaweicloud_vpc_subnet" "public" {
  name       = "${local.name_prefix}-public-subnet"
  cidr       = var.public_subnets[0]
  gateway_ip = cidrhost(var.public_subnets[0], 1)
  vpc_id     = huaweicloud_vpc.main.id
  
  dns_list = ["100.125.1.250", "100.125.21.250"]
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-public-subnet"
    Type = "public"
  })
}

# Security Group
resource "huaweicloud_networking_secgroup" "main" {
  name        = "${local.name_prefix}-sg"
  description = "Security group for PulseExpends ECS instances"
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-sg"
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# Security Group Rules
resource "huaweicloud_networking_secgroup_rule" "ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "mcp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8080
  port_range_max    = 8080
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "pdf_parser" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8000
  port_range_max    = 8000
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "auth_server" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8082
  port_range_max    = 8082
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "openclaw_dashboard" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 18789
  port_range_max    = 18789
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "egress_all" {
  direction         = "egress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

# EIP for Domain
resource "huaweicloud_vpc_eip" "domain" {
  count = var.enable_domain ? 1 : 0
  
  publicip {
    type = "5_bgp"
  }
  
  bandwidth {
    name        = "${local.name_prefix}-domain-bandwidth"
    size        = var.domain_bandwidth_size
    share_type  = "PER"
    charge_mode = "traffic"
  }
  
  tags = merge(local.common_tags, {
    Name    = "${local.name_prefix}-domain-eip"
    Purpose = "Domain"
    Domain  = var.domain_name
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# ECS Instance
resource "huaweicloud_compute_instance" "main" {
  name              = "${local.name_prefix}-ecs"
  flavor_name       = var.ecs_instance_type
  availability_zone = "la-south-2a"
  key_pair          = var.ecs_key_pair
  
  security_groups = [huaweicloud_networking_secgroup.main.name]
  
  network {
    uuid = huaweicloud_vpc_subnet.public.id
  }
  
  system_disk_type = "SAS"
  system_disk_size = 50
  
  data_disks {
    type = "SAS"
    size = 100
  }
  
  image_name = "Ubuntu 22.04 server 64bit"
  
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
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-ecs"
    Role = "Application Server"
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# Associate EIP with ECS if domain is enabled
resource "huaweicloud_compute_eip_associate" "domain_association" {
  count = var.enable_domain ? 1 : 0
  
  public_ip   = huaweicloud_vpc_eip.domain[0].address
  instance_id = huaweicloud_compute_instance.main.id
}

# OBS Bucket for Data
resource "huaweicloud_obs_bucket" "data" {
  bucket = "${var.project_name}-data-${var.environment}-${random_id.deployment.hex}"
  acl    = "private"
  
  # Versioning
  versioning = var.obs_versioning_enabled
  
  # Server-side encryption
  # Note: KMS encryption requires manual configuration in OBS console
  # or using OBS API with specific headers
  
  tags = merge(local.common_tags, {
    Name    = "${local.name_prefix}-obs-data"
    Purpose = "Application Data"
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# OBS Bucket for Documents
resource "huaweicloud_obs_bucket" "documents" {
  bucket = "${var.project_name}-documents-${var.environment}-${random_id.deployment.hex}"
  acl    = "private"
  
  # Versioning
  versioning = var.obs_versioning_enabled
  
  # Server-side encryption
  # Note: KMS encryption requires manual configuration in OBS console
  # or using OBS API with specific headers
  
  tags = merge(local.common_tags, {
    Name    = "${local.name_prefix}-obs-documents"
    Purpose = "Document Storage"
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# KMS Key for Encryption
resource "huaweicloud_kms_key" "main" {
  count = var.enable_kms ? 1 : 0
  
  key_alias       = var.kms_key_alias
  key_description = "KMS key for PulseExpends ${var.environment} environment"
  
  pending_days = 7
  is_enabled   = true
  
  tags = merge(local.common_tags, {
    Name        = "${local.name_prefix}-kms-key"
    Purpose     = "Encryption"
    Environment = var.environment
  })
  
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# Outputs
output "enterprise_project_id" {
  description = "ID of the Enterprise Project"
  value       = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = huaweicloud_vpc.main.id
}

output "public_subnet_id" {
  description = "ID of the public subnet"
  value       = huaweicloud_vpc_subnet.public.id
}

output "security_group_id" {
  description = "ID of the security group"
  value       = huaweicloud_networking_secgroup.main.id
}

output "ecs_instance_id" {
  description = "ID of the ECS instance"
  value       = huaweicloud_compute_instance.main.id
}

output "ecs_public_ip" {
  description = "Public IP address of the ECS instance"
  value       = var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : huaweicloud_compute_instance.main.access_ip_v4
}

output "ecs_private_ip" {
  description = "Private IP address of the ECS instance"
  value       = huaweicloud_compute_instance.main.access_ip_v4
}

output "domain_ip_address" {
  description = "Public IP address for the domain"
  value       = var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : null
}

output "data_bucket_name" {
  description = "Name of the OBS data bucket"
  value       = huaweicloud_obs_bucket.data.bucket
}

output "documents_bucket_name" {
  description = "Name of the OBS documents bucket"
  value       = huaweicloud_obs_bucket.documents.bucket
}

output "kms_key_id" {
  description = "ID of the KMS key"
  value       = var.enable_kms ? huaweicloud_kms_key.main[0].id : null
}

output "ssh_connection_command" {
  description = "SSH connection command"
  value       = "ssh -i pulse-expends-key.pem ubuntu@${var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : huaweicloud_compute_instance.main.access_ip_v4}"
}

output "application_url" {
  description = "Application URL"
  value       = var.enable_domain ? "http://${var.domain_name}" : "http://${var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : huaweicloud_compute_instance.main.access_ip_v4}"
}

output "mcp_server_url" {
  description = "MCP Server URL"
  value       = var.enable_domain ? "http://${var.domain_name}:8080" : "http://${var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : huaweicloud_compute_instance.main.access_ip_v4}:8080"
}

output "pdf_parser_url" {
  description = "PDF Parser Service URL"
  value       = var.enable_domain ? "http://${var.domain_name}:8000" : "http://${var.enable_domain ? huaweicloud_vpc_eip.domain[0].address : huaweicloud_compute_instance.main.access_ip_v4}:8000"
}