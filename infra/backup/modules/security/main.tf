# Security Module for PulseExpends

# Create IAM users (disabled for now - configure manually if needed)
# resource "huaweicloud_identity_user" "main" {
#   for_each = { for user in var.iam_users : user.name => user }
#   
#   name     = each.value.name
#   enabled  = each.value.enabled
#   password = each.value.password
#   
#   description = each.value.description
#   
#   tags = merge(var.common_tags, {
#     Name        = each.value.name
#     Role        = each.value.role
#     Environment = var.environment
#   })
# }
# 
# # Create IAM groups
# resource "huaweicloud_identity_group" "main" {
#   for_each = { for group in var.iam_groups : group.name => group }
#   
#   name        = each.value.name
#   description = each.value.description
#   
#   tags = merge(var.common_tags, {
#     Name        = each.value.name
#     Environment = var.environment
#   })
# }
# 
# # Assign users to groups
# resource "huaweicloud_identity_group_membership" "main" {
#   for_each = { for assignment in local.group_assignments : "${assignment.user}.${assignment.group}" => assignment }
#   
#   group = each.value.group
#   users = [each.value.user]
# }
# 
# # Create IAM policies
# resource "huaweicloud_identity_role" "custom" {
#   for_each = { for policy in var.iam_policies : policy.name => policy }
#   
#   name        = each.value.name
#   description = each.value.description
#   type        = "AX"  # Custom policy
#   
#   policy = each.value.policy_document
#   
#   tags = merge(var.common_tags, {
#     Name        = each.value.name
#     Environment = var.environment
#   })
# }
# 
# # Attach policies to groups
# resource "huaweicloud_identity_role_assignment" "group" {
#   for_each = { for assignment in local.policy_group_assignments : "${assignment.policy}.${assignment.group}" => assignment }
#   
#   role_id   = each.value.policy_id
#   group_id  = each.value.group_id
#   project_id = var.project_id
# }
# 
# # Attach policies to users
# resource "huaweicloud_identity_role_assignment" "user" {
#   for_each = { for assignment in local.policy_user_assignments : "${assignment.policy}.${assignment.user}" => assignment }
#   
#   role_id   = each.value.policy_id
#   user_id   = each.value.user_id
#   project_id = var.project_id
# }

# Create KMS key if enabled
resource "huaweicloud_kms_key" "main" {
  count = var.enable_kms ? 1 : 0
  
  key_alias       = var.kms_key_alias
  key_description = "KMS key for PulseExpends ${var.environment} environment"
  realm           = var.region
  
  pending_days = 7
  is_enabled   = true
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-kms-key"
    Purpose     = "Encryption"
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create WAF policy if enabled
resource "huaweicloud_waf_policy" "main" {
  count = var.enable_waf ? 1 : 0
  
  name = var.waf_policy_name
  
  protection_mode = "block"
  level          = 2  # Medium
  
  dynamic "robot_action" {
    for_each = ["crawler", "scanner", "script", "other"]
    
    content {
      category = robot_action.value
      action   = "block"
    }
  }
  
  tags = merge(var.common_tags, {
    Name        = var.waf_policy_name
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Create Anti-DDoS protection if enabled
resource "huaweicloud_antiddos" "main" {
  count = var.enable_antiddos ? 1 : 0
  
  floating_ip_id = var.eip_id
  
  enable_l7   = true
  traffic_pos_id = 1
  http_request_pos_id = 2
  cleaning_access_pos_id = 3
  app_type_id = 0  # Web application
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-antiddos"
    Environment = var.environment
  })
  
  enterprise_project_id = var.enterprise_project_id
}

# Local variables for user-group assignments (disabled for now)
locals {
  # Empty locals for now since IAM is disabled
}

# Outputs
output "kms_key_id" {
  description = "ID of the KMS key"
  value       = var.enable_kms ? huaweicloud_kms_key.main[0].id : null
}

output "kms_key_alias" {
  description = "Alias of the KMS key"
  value       = var.enable_kms ? huaweicloud_kms_key.main[0].key_alias : null
}

output "waf_policy_id" {
  description = "ID of the WAF policy"
  value       = var.enable_waf ? huaweicloud_waf_policy.main[0].id : null
}

output "antiddos_id" {
  description = "ID of the Anti-DDoS protection"
  value       = var.enable_antiddos ? huaweicloud_antiddos.main[0].id : null
}