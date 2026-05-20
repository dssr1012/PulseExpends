# Network Module for PulseExpends

# Create VPC
resource "huaweicloud_vpc" "main" {
  name = "${var.name_prefix}-vpc"
  cidr = var.vpc_cidr
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-vpc"
    Type = "VPC"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create public subnets
resource "huaweicloud_vpc_subnet" "public" {
  count = length(var.public_subnets)
  
  name       = "${var.name_prefix}-public-${count.index + 1}"
  cidr       = var.public_subnets[count.index]
  gateway_ip = cidrhost(var.public_subnets[count.index], 1)
  vpc_id     = huaweicloud_vpc.main.id
  
  dns_list = ["100.125.1.250", "100.125.21.250"]  # Huawei Cloud DNS
  
  tags = merge(var.common_tags, {
    Name     = "${var.name_prefix}-public-${count.index + 1}"
    Type     = "Public"
    Zone     = element(data.huaweicloud_availability_zones.available.names, count.index % length(data.huaweicloud_availability_zones.available.names))
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create private subnets
resource "huaweicloud_vpc_subnet" "private" {
  count = length(var.private_subnets)
  
  name       = "${var.name_prefix}-private-${count.index + 1}"
  cidr       = var.private_subnets[count.index]
  gateway_ip = cidrhost(var.private_subnets[count.index], 1)
  vpc_id     = huaweicloud_vpc.main.id
  
  dns_list = ["100.125.1.250", "100.125.21.250"]  # Huawei Cloud DNS
  
  tags = merge(var.common_tags, {
    Name     = "${var.name_prefix}-private-${count.index + 1}"
    Type     = "Private"
    Zone     = element(data.huaweicloud_availability_zones.available.names, count.index % length(data.huaweicloud_availability_zones.available.names))
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create security groups
resource "huaweicloud_networking_secgroup" "main" {
  name        = "${var.name_prefix}-sg"
  description = "Security group for PulseExpends application"
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-sg"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Security group rules for web access
resource "huaweicloud_networking_secgroup_rule" "web_http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "web_https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

# Security group rules for application ports
resource "huaweicloud_networking_secgroup_rule" "app_mcp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8080
  port_range_max    = 8080
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

resource "huaweicloud_networking_secgroup_rule" "app_python" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 8000
  port_range_max    = 8000
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

# SSH access
resource "huaweicloud_networking_secgroup_rule" "ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

# Outbound traffic
resource "huaweicloud_networking_secgroup_rule" "egress_all" {
  direction         = "egress"
  ethertype         = "IPv4"
  protocol          = ""
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = huaweicloud_networking_secgroup.main.id
}

# Create NAT Gateway if enabled
resource "huaweicloud_nat_gateway" "main" {
  count = var.enable_nat_gateway ? 1 : 0
  
  name        = "${var.name_prefix}-nat"
  spec        = "1"
  vpc_id      = huaweicloud_vpc.main.id
  subnet_id   = huaweicloud_vpc_subnet.public[0].id
  description = "NAT Gateway for private subnet internet access"
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-nat"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create SNAT rule for NAT Gateway
resource "huaweicloud_nat_snat_rule" "main" {
  count = var.enable_nat_gateway ? 1 : 0
  
  nat_gateway_id = huaweicloud_nat_gateway.main[0].id
  subnet_id      = huaweicloud_vpc_subnet.private[0].id
  floating_ip_id = huaweicloud_vpc_eip.nat[0].id
  
  description = "SNAT rule for private subnet"
}

# Create EIP for NAT Gateway
resource "huaweicloud_vpc_eip" "nat" {
  count = var.enable_nat_gateway ? 1 : 0
  
  publicip {
    type = "5_bgp"
  }
  
  bandwidth {
    name        = "${var.name_prefix}-nat-bandwidth"
    size        = 5  # 5 Mbps
    share_type  = "PER"
    charge_mode = "bandwidth"
  }
  
  tags = merge(var.common_tags, {
    Name    = "${var.name_prefix}-nat-eip"
    Purpose = "NAT Gateway"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# VPC Flow Logs if enabled
resource "huaweicloud_vpc_flow_log" "main" {
  count = var.enable_flow_logs ? 1 : 0
  
  name         = "${var.name_prefix}-flow-log"
  resource_type = "vpc"
  resource_id  = huaweicloud_vpc.main.id
  log_group_id = huaweicloud_lts_group.main[0].id
  log_stream_id = huaweicloud_lts_stream.main[0].id
  traffic_type = "all"
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-flow-log"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# LTS Group for Flow Logs
resource "huaweicloud_lts_group" "main" {
  count = var.enable_flow_logs ? 1 : 0
  
  group_name  = "${var.name_prefix}-lts-group"
  ttl_in_days = 7
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-lts-group"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# LTS Stream for Flow Logs
resource "huaweicloud_lts_stream" "main" {
  count = var.enable_flow_logs ? 1 : 0
  
  group_id    = huaweicloud_lts_group.main[0].id
  stream_name = "${var.name_prefix}-flow-log-stream"
  
  enterprise_project_id = var.enterprise_project_id
}

# Data source for availability zones
data "huaweicloud_availability_zones" "available" {}

# Outputs
output "vpc_id" {
  description = "ID of the created VPC"
  value       = huaweicloud_vpc.main.id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = huaweicloud_vpc_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = huaweicloud_vpc_subnet.private[*].id
}

output "security_group_ids" {
  description = "IDs of the created security groups"
  value       = [huaweicloud_networking_secgroup.main.id]
}

output "nat_gateway_id" {
  description = "ID of the NAT Gateway"
  value       = var.enable_nat_gateway ? huaweicloud_nat_gateway.main[0].id : null
}

output "nat_eip_address" {
  description = "EIP address of the NAT Gateway"
  value       = var.enable_nat_gateway ? huaweicloud_vpc_eip.nat[0].address : null
}