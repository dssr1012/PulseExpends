# Compute Module for PulseExpends

# Create ECS instances
resource "huaweicloud_compute_instance" "main" {
  count = var.ecs_instance_count
  
  name              = "${var.name_prefix}-ecs-${count.index + 1}"
  flavor_id         = var.ecs_instance_type
  availability_zone = element(data.huaweicloud_availability_zones.available.names, count.index % length(data.huaweicloud_availability_zones.available.names))
  key_pair          = var.ecs_key_pair
  
  security_group_ids = var.security_group_ids
  
  network {
    uuid = element(var.public_subnet_ids, count.index % length(var.public_subnet_ids))
  }
  
  system_disk_type = "SAS"
  system_disk_size = 50  # GB
  
  data_disks {
    type = "SAS"
    size = 100  # GB
  }
  
  image_id = var.ecs_image_id
  
  user_data = var.user_data
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecs-${count.index + 1}"
    Role        = "application"
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create EIPs for ECS instances if not using load balancer
resource "huaweicloud_vpc_eip" "ecs" {
  count = var.enable_load_balancer ? 0 : var.ecs_instance_count
  
  publicip {
    type = "5_bgp"
  }
  
  bandwidth {
    name        = "${var.name_prefix}-ecs-${count.index + 1}-bandwidth"
    size        = 5  # Mbps
    share_type  = "PER"
    charge_mode = "bandwidth"
  }
  
  tags = merge(var.common_tags, {
    Name    = "${var.name_prefix}-ecs-${count.index + 1}-eip"
    Purpose = "ECS Public IP"
    Instance = huaweicloud_compute_instance.main[count.index].name
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Associate EIPs with ECS instances
resource "huaweicloud_compute_eip_associate" "main" {
  count = var.enable_load_balancer ? 0 : var.ecs_instance_count
  
  public_ip   = huaweicloud_vpc_eip.ecs[count.index].address
  instance_id = huaweicloud_compute_instance.main[count.index].id
}

# Create load balancer if enabled
resource "huaweicloud_elb_loadbalancer" "main" {
  count = var.enable_load_balancer ? 1 : 0
  
  name          = "${var.name_prefix}-elb"
  description   = "Load balancer for PulseExpends application"
  vpc_id        = var.vpc_id
  ipv4_subnet_id = element(var.public_subnet_ids, 0)
  
  l4_flavor_id = "L4_flavor.elb.s2.small"
  l7_flavor_id = "L7_flavor.elb.s2.small"
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-elb"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create listener for load balancer
resource "huaweicloud_elb_listener" "http" {
  count = var.enable_load_balancer ? 1 : 0
  
  name            = "${var.name_prefix}-http-listener"
  description     = "HTTP listener for PulseExpends"
  protocol        = var.lb_protocol
  protocol_port   = var.lb_listener_port
  loadbalancer_id = huaweicloud_elb_loadbalancer.main[0].id
  
  idle_timeout     = 60
  request_timeout  = 60
  response_timeout = 60
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-http-listener"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create pool for load balancer
resource "huaweicloud_elb_pool" "main" {
  count = var.enable_load_balancer ? 1 : 0
  
  name        = "${var.name_prefix}-pool"
  description = "Backend pool for PulseExpends"
  protocol    = var.lb_protocol
  lb_method   = "ROUND_ROBIN"
  listener_id = huaweicloud_elb_listener.http[0].id
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-pool"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Add ECS instances to load balancer pool
resource "huaweicloud_elb_member" "main" {
  count = var.enable_load_balancer ? var.ecs_instance_count : 0
  
  address       = huaweicloud_compute_instance.main[count.index].access_ip_v4
  protocol_port = 8080
  subnet_id     = element(var.private_subnet_ids, count.index % length(var.private_subnet_ids))
  pool_id       = huaweicloud_elb_pool.main[0].id
  
  enterprise_project_id = var.enterprise_project_id
}

# Create health monitor for load balancer (disabled - resource type not supported)
# resource "huaweicloud_elb_healthcheck" "main" {
#   count = var.enable_load_balancer ? 1 : 0
#   
#   protocol    = "TCP"
#   delay       = 5
#   timeout     = 3
#   max_retries = 3
#   pool_id     = huaweicloud_elb_pool.main[0].id
#   
#   enterprise_project_id = var.enterprise_project_id
# }

# Create auto scaling group if enabled
resource "huaweicloud_as_group" "main" {
  count = var.enable_auto_scaling ? 1 : 0
  
  scaling_group_name       = "${var.name_prefix}-asg"
  scaling_configuration_id = huaweicloud_as_configuration.main[0].id
  desire_instance_number   = var.desired_capacity
  min_instance_number      = var.min_size
  max_instance_number      = var.max_size
  cool_down_time          = 300
  health_periodic_audit_time = 5
  health_periodic_audit_grace_period = 600
  
  networks {
    id = element(var.private_subnet_ids, 0)
  }
  
  security_groups {
    id = element(var.security_group_ids, 0)
  }
  
  lbaas_listeners {
    pool_id       = huaweicloud_elb_pool.main[0].id
    protocol_port = 8080
  }
  
  tags = merge(var.common_tags, {
    Name = "${var.name_prefix}-asg"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create auto scaling configuration
resource "huaweicloud_as_configuration" "main" {
  count = var.enable_auto_scaling ? 1 : 0
  
  scaling_configuration_name = "${var.name_prefix}-as-config"
  instance_config {
    flavor   = var.ecs_instance_type
    image    = var.ecs_image_id
    key_name = var.ecs_key_pair
    
    disk {
      size        = 50
      volume_type = "SAS"
      disk_type   = "SYS"
    }
    
    disk {
      size        = 100
      volume_type = "SAS"
      disk_type   = "DATA"
    }
    
    user_data = var.user_data
  }
  
  # enterprise_project_id = var.enterprise_project_id  # Not supported for this resource
}

# Data source for availability zones
data "huaweicloud_availability_zones" "available" {}

# Outputs
output "ecs_instance_ids" {
  description = "IDs of the ECS instances"
  value       = huaweicloud_compute_instance.main[*].id
}

output "ecs_public_ips" {
  description = "Public IP addresses of the ECS instances"
  value       = var.enable_load_balancer ? [] : huaweicloud_vpc_eip.ecs[*].address
}

output "ecs_private_ips" {
  description = "Private IP addresses of the ECS instances"
  value       = huaweicloud_compute_instance.main[*].access_ip_v4
}

output "load_balancer_id" {
  description = "ID of the load balancer"
  value       = var.enable_load_balancer ? huaweicloud_elb_loadbalancer.main[0].id : null
}

output "load_balancer_ip" {
  description = "Public IP address of the load balancer"
  value       = var.enable_load_balancer ? huaweicloud_elb_loadbalancer.main[0].ipv4_address : null
}

output "as_group_id" {
  description = "ID of the auto scaling group"
  value       = var.enable_auto_scaling ? huaweicloud_as_group.main[0].id : null
}