# ============================================================================
# Enterprise Project Configuration
# ============================================================================

# Create Enterprise Project
resource "huaweicloud_enterprise_project" "pulse_expends" {
  count = var.enable_enterprise_project ? 1 : 0
  
  name        = var.enterprise_project_name
  description = "Enterprise Project for PulseExpends application"
  type        = var.enterprise_project_type
}

# Create EIP for domain if enabled
resource "huaweicloud_vpc_eip" "domain" {
  count = var.enable_domain ? 1 : 0
  
  publicip {
    type = "5_bgp"
  }
  
  bandwidth {
    name        = "${local.name_prefix}-domain-bandwidth"
    size        = var.domain_bandwidth_size
    share_type  = "PER"
    charge_mode = "bandwidth"
  }
  
  tags = merge(local.common_tags, {
    Name        = "${local.name_prefix}-domain-eip"
    Purpose     = "Domain"
    Domain      = var.domain_name
  })
  
  # Associate with enterprise project if enabled
  enterprise_project_id = var.enable_enterprise_project ? huaweicloud_enterprise_project.pulse_expends[0].id : null
}

# DNS Record for domain (using Huawei Cloud DNS)
resource "huaweicloud_dns_recordset" "pulseexpends" {
  count = var.enable_domain && var.create_dns_record && var.dns_zone_id != "" ? 1 : 0
  
  zone_id = var.dns_zone_id
  name    = var.domain_name
  type    = "A"
  ttl     = 300
  
  records = [huaweicloud_vpc_eip.domain[0].address]
  
  tags = merge(local.common_tags, {
    Name    = "${local.name_prefix}-dns-record"
    Domain  = var.domain_name
    Type    = "A"
  })
}

# SSL Certificate for domain (optional)
resource "huaweicloud_waf_certificate" "ssl" {
  count = var.enable_domain && var.enable_ssl && var.ssl_certificate != "" && var.ssl_private_key != "" ? 1 : 0
  
  name = "${local.name_prefix}-ssl-cert"
  
  certificate = var.ssl_certificate
  private_key = var.ssl_private_key
}