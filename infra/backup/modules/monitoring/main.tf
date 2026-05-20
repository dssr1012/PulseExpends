# Monitoring Module for PulseExpends

# Create Cloud Eye alarm rule for ECS CPU usage
resource "huaweicloud_ces_alarmrule" "ecs_cpu" {
  count = length(var.ecs_instance_ids) > 0 ? 1 : 0
  
  alarm_name        = "${var.name_prefix}-ecs-cpu-alarm"
  alarm_description = "ECS CPU usage alarm"
  alarm_level       = 2  # Critical
  
  metric {
    namespace   = "SYS.ECS"
    metric_name = "cpu_util"
    dimensions {
      name  = "instance_id"
      value = var.ecs_instance_ids[0]
    }
  }
  
  condition {
    period              = 300
    filter             = "average"
    comparison_operator = ">"
    value              = 80
    unit               = "%"
    count              = 1
  }
  
  alarm_actions {
    type              = "notification"
    notification_list = var.alert_recipients
  }
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecs-cpu-alarm"
    Metric      = "cpu_util"
    Resource    = "ECS"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create Cloud Eye alarm rule for ECS memory usage
resource "huaweicloud_ces_alarmrule" "ecs_memory" {
  count = length(var.ecs_instance_ids) > 0 ? 1 : 0
  
  alarm_name        = "${var.name_prefix}-ecs-memory-alarm"
  alarm_description = "ECS memory usage alarm"
  alarm_level       = 2  # Critical
  
  metric {
    namespace   = "SYS.ECS"
    metric_name = "mem_util"
    dimensions {
      name  = "instance_id"
      value = var.ecs_instance_ids[0]
    }
  }
  
  condition {
    period              = 300
    filter             = "average"
    comparison_operator = ">"
    value              = 85
    unit               = "%"
    count              = 1
  }
  
  alarm_actions {
    type              = "notification"
    notification_list = var.alert_recipients
  }
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecs-memory-alarm"
    Metric      = "mem_util"
    Resource    = "ECS"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create Cloud Eye alarm rule for ECS disk usage
resource "huaweicloud_ces_alarmrule" "ecs_disk" {
  count = length(var.ecs_instance_ids) > 0 ? 1 : 0
  
  alarm_name        = "${var.name_prefix}-ecs-disk-alarm"
  alarm_description = "ECS disk usage alarm"
  alarm_level       = 2  # Critical
  
  metric {
    namespace   = "SYS.ECS"
    metric_name = "disk_util_inband"
    dimensions {
      name  = "instance_id"
      value = var.ecs_instance_ids[0]
    }
  }
  
  condition {
    period              = 300
    filter             = "average"
    comparison_operator = ">"
    value              = 90
    unit               = "%"
    count              = 1
  }
  
  alarm_actions {
    type              = "notification"
    notification_list = var.alert_recipients
  }
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecs-disk-alarm"
    Metric      = "disk_util_inband"
    Resource    = "ECS"
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create Cloud Eye dashboard if enabled (disabled for now - API issues)
# resource "huaweicloud_ces_dashboard" "main" {
#   count = var.enable_dashboards ? 1 : 0
#   
#   dashboard_name = "${var.name_prefix}-dashboard"
#   
#   # ECS CPU widget
#   widget {
#     x      = 0
#     y      = 0
#     width  = 12
#     height = 6
#     
#     properties {
#       title = "ECS CPU Usage"
#       type  = "line"
#       
#       metric {
#         namespace   = "SYS.ECS"
#         metric_name = "cpu_util"
#         dimensions {
#           name  = "instance_id"
#           value = length(var.ecs_instance_ids) > 0 ? var.ecs_instance_ids[0] : ""
#         }
#       }
#     }
#   }
#   
#   # ECS Memory widget
#   widget {
#     x      = 12
#     y      = 0
#     width  = 12
#     height = 6
#     
#     properties {
#       title = "ECS Memory Usage"
#       type  = "line"
#       
#       metric {
#         namespace   = "SYS.ECS"
#         metric_name = "mem_util"
#         dimensions {
#           name  = "instance_id"
#           value = length(var.ecs_instance_ids) > 0 ? var.ecs_instance_ids[0] : ""
#         }
#       }
#     }
#   }
#   
#   # ECS Disk widget
#   widget {
#     x      = 0
#     y      = 6
#     width  = 12
#     height = 6
#     
#     properties {
#       title = "ECS Disk Usage"
#       type  = "line"
#       
#       metric {
#         namespace   = "SYS.ECS"
#         metric_name = "disk_util_inband"
#         dimensions {
#           name  = "instance_id"
#           value = length(var.ecs_instance_ids) > 0 ? var.ecs_instance_ids[0] : ""
#         }
#       }
#     }
#   }
#   
#   # Load Balancer widget (if enabled)
#   widget {
#     x      = 12
#     y      = 6
#     width  = 12
#     height = 6
#     
#     properties {
#       title = "Load Balancer Connections"
#       type  = "line"
#       
#       metric {
#         namespace   = "SYS.ELB"
#         metric_name = "m1_cps"
#         dimensions {
#           name  = "lb_instance_id"
#           value = var.load_balancer_id != null ? var.load_balancer_id : ""
#         }
#       }
#     }
#   }
#   
#   tags = merge(var.common_tags, {
#     Name        = "${var.name_prefix}-dashboard"
#     Environment = var.environment
#   })
#   
#   enterprise_project_id = var.enterprise_project_id
# }

# Outputs
# output "dashboard_url" {
#   description = "URL of the Cloud Eye dashboard"
#   value       = var.enable_dashboards ? huaweicloud_ces_dashboard.main[0].dashboard_url : null
# }

output "alarm_rule_ids" {
  description = "IDs of the created alarm rules"
  value = {
    ecs_cpu    = length(huaweicloud_ces_alarmrule.ecs_cpu) > 0 ? huaweicloud_ces_alarmrule.ecs_cpu[0].id : null
    ecs_memory = length(huaweicloud_ces_alarmrule.ecs_memory) > 0 ? huaweicloud_ces_alarmrule.ecs_memory[0].id : null
    ecs_disk   = length(huaweicloud_ces_alarmrule.ecs_disk) > 0 ? huaweicloud_ces_alarmrule.ecs_disk[0].id : null
  }
}