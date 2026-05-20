# Variables for Monitoring Module

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "region" {
  description = "Huawei Cloud region"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

variable "enterprise_project_id" {
  description = "Enterprise Project ID for resource association"
  type        = string
  default     = null
}

variable "ecs_instance_ids" {
  description = "List of ECS instance IDs to monitor"
  type        = list(string)
  default     = []
}

variable "load_balancer_id" {
  description = "Load balancer ID to monitor"
  type        = string
  default     = null
}

variable "database_id" {
  description = "Database ID to monitor"
  type        = string
  default     = null
}

variable "redis_id" {
  description = "Redis instance ID to monitor"
  type        = string
  default     = null
}

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
  description = "Enable Cloud Eye dashboard creation"
  type        = bool
  default     = true
}