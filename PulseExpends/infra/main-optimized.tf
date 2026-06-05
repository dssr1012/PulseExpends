# PulseExpends Infrastructure - Optimized Architecture
# Single ECS with EIP in Public Subnet, RDS in Private Subnet
# No NAT Gateway - ECS has direct internet access via EIP

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

# ============================================================================
# VPC and Networking
# ============================================================================

# VPC
resource "huaweicloud_vpc" "main" {
  name = "${local.name_prefix}-vpc"
  cidr = var.vpc_cidr
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-vpc"
  })
  
  enterprise_project_id = "9d731c5b-e130-430b-b88d-66f312596926"
}

# Public Subnet for ECS (with internet access via EIP)
resource "huaweicloud_vpc_subnet" "ecs_subnet" {
  name       = "${local.name_prefix}-ecs-subnet"
  cidr       = var.public_subnets[0]
  gateway_ip = cidrhost(var.public_subnets[0], 1)
  vpc_id     = huaweicloud_vpc.main.id
  
  dns_list = ["100.125.1.250", "100.125.21.250"]
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-ecs-subnet"
    Type = "public"
  })
}

# Private Subnet for RDS (no internet access)
resource "huaweicloud_vpc_subnet" "rds_subnet" {
  name       = "${local.name_prefix}-rds-subnet"
  cidr       = var.private_subnets[0]
  gateway_ip = cidrhost(var.private_subnets[0], 1)
  vpc_id     = huaweicloud_vpc.main.id
  
  dns_list = ["100.125.1.250", "100.125.21.250"]
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-rds-subnet"
    Type = "private"
  })
}

# ============================================================================
# Security Groups
# ============================================================================

# Security Group for ECS (allows inbound from internet for services)
resource "huaweicloud_networking_secgroup" "ecs_sg" {
  name        = "${local.name_prefix}-ecs-sg"
  description = "Security group for PulseExpends ECS instance"
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-ecs-sg"
  })
  
  enterprise_project_id = "9d731c5b-e130-430b-b88d-66f312596926"
}

# Security Group for RDS (only allows traffic from ECS security group)
resource "huaweicloud_networking_secgroup" "rds_sg" {
  name        = "${local.name_prefix}-rds-sg"
  description = "Security group for PulseExpends RDS database"
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-rds-sg"
  })
  
  enterprise_project_id = "9d731c5b-e130-430b-b88d-66f312596926"
}

# ECS Security Group Rules
resource "huaweicloud_networking_secgroup_rule" "ecs_ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "SSH access"
}

resource "huaweicloud_networking_secgroup_rule" "ecs_http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "HTTP access"
}

resource "huaweicloud_networking_secgroup_rule" "ecs_https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "HTTPS access"
}

resource "huaweicloud_networking_secgroup_rule" "ecs_mcp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8080
  port_range_max    = 8080
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "MCP server access"
}

resource "huaweicloud_networking_secgroup_rule" "ecs_pdf_parser" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8000
  port_range_max    = 8000
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "PDF parser service access"
}

resource "huaweicloud_networking_secgroup_rule" "ecs_auth_server" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8082
  port_range_max    = 8082
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "Auth server access"
}

resource "huaweicloud_networking_secgroup_rule" "ecs_egress_all" {
  direction         = "egress"
  ethertype         = "IPv4"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.ecs_sg.id
  description       = "Allow all outbound traffic"
}

# RDS Security Group Rule - ONLY allow PostgreSQL from ECS security group
resource "huaweicloud_networking_secgroup_rule" "rds_postgresql" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 5432
  port_range_max    = 5432
  remote_group_id   = huaweicloud_networking_secgroup.ecs_sg.id  # Critical: Reference ECS SG, not CIDR
  security_group_id = huaweicloud_networking_secgroup.rds_sg.id
  description       = "Allow PostgreSQL access from ECS instances"
}

resource "huaweicloud_networking_secgroup_rule" "rds_egress_all" {
  direction         = "egress"
  ethertype         = "IPv4"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.rds_sg.id
  description       = "Allow all outbound traffic from RDS"
}

# ============================================================================
# Elastic IP for ECS
# ============================================================================

resource "huaweicloud_vpc_eip" "ecs_eip" {
  name = "${local.name_prefix}-ecs-eip"
  
  publicip {
    type = "5_bgp"
  }
  
  bandwidth {
    name        = "${local.name_prefix}-ecs-bandwidth"
    size        = var.eip_bandwidth_size
    share_type  = "PER"
    charge_mode = "traffic"
  }
  
  tags = merge(local.common_tags, {
    Name    = "${local.name_prefix}-ecs-eip"
    Purpose = "ECS Public IP"
  })
  
  enterprise_project_id = "9d731c5b-e130-430b-b88d-66f312596926"
}

# ============================================================================
# ECS Instance
# ============================================================================

resource "huaweicloud_compute_instance" "main" {
  name              = "${local.name_prefix}-ecs"
  flavor_name       = var.ecs_instance_type
  availability_zone = var.availability_zone
  key_pair          = var.ecs_key_pair
  
  security_groups = [huaweicloud_networking_secgroup.ecs_sg.name]
  
  network {
    uuid = huaweicloud_vpc_subnet.ecs_subnet.id
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
  
  enterprise_project_id = "9d731c5b-e130-430b-b88d-66f312596926"
}

# Associate EIP with ECS
resource "huaweicloud_compute_eip_associate" "ecs_eip_association" {
  public_ip   = huaweicloud_vpc_eip.ecs_eip.address
  instance_id = huaweicloud_compute_instance.main.id
}

# ============================================================================
# Outputs
# ============================================================================

output "vpc_id" {
  description = "ID of the VPC"
  value       = huaweicloud_vpc.main.id
}

output "ecs_subnet_id" {
  description = "ID of the ECS subnet"
  value       = huaweicloud_vpc_subnet.ecs_subnet.id
}

output "rds_subnet_id" {
  description = "ID of the RDS subnet"
  value       = huaweicloud_vpc_subnet.rds_subnet.id
}

output "ecs_security_group_id" {
  description = "ID of the ECS security group"
  value       = huaweicloud_networking_secgroup.ecs_sg.id
}

output "rds_security_group_id" {
  description = "ID of the RDS security group"
  value       = huaweicloud_networking_secgroup.rds_sg.id
}

output "ecs_instance_id" {
  description = "ID of the ECS instance"
  value       = huaweicloud_compute_instance.main.id
}

output "ecs_public_ip" {
  description = "Public IP address of the ECS instance"
  value       = huaweicloud_vpc_eip.ecs_eip.address
}

output "ecs_private_ip" {
  description = "Private IP address of the ECS instance"
  value       = huaweicloud_compute_instance.main.access_ip_v4
}

output "eip_address" {
  description = "Elastic IP address"
  value       = huaweicloud_vpc_eip.ecs_eip.address
}

output "ssh_connection_command" {
  description = "SSH connection command"
  value       = "ssh -i ${var.ecs_key_pair}.pem ubuntu@${huaweicloud_vpc_eip.ecs_eip.address}"
}

output "application_url" {
  description = "Application URL"
  value       = "http://${huaweicloud_vpc_eip.ecs_eip.address}"
}

output "mcp_server_url" {
  description = "MCP Server URL"
  value       = "http://${huaweicloud_vpc_eip.ecs_eip.address}:8080"
}

output "pdf_parser_url" {
  description = "PDF Parser Service URL"
  value       = "http://${huaweicloud_vpc_eip.ecs_eip.address}:8000"
}

output "auth_server_url" {
  description = "Auth Server URL"
  value       = "http://${huaweicloud_vpc_eip.ecs_eip.address}:8082"
}