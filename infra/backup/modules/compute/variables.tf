# Variables for Compute Module

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

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs"
  type        = list(string)
}

variable "ecs_instance_type" {
  description = "ECS instance type"
  type        = string
  default     = "c6.large.2"
}

variable "ecs_image_id" {
  description = "ECS image ID"
  type        = string
  default     = ""  # Will be set based on region
}

variable "ecs_key_pair" {
  description = "ECS key pair name"
  type        = string
  default     = "pulse-expends-key"
}

variable "ecs_instance_count" {
  description = "Number of ECS instances"
  type        = number
  default     = 1
}

variable "enable_auto_scaling" {
  description = "Enable auto scaling"
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
  description = "Enable load balancer"
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
}

variable "enable_domain" {
  description = "Enable domain configuration"
  type        = bool
  default     = false
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = ""
}

variable "domain_ip_address" {
  description = "Domain IP address (EIP)"
  type        = string
  default     = null
}

variable "enable_ssl" {
  description = "Enable SSL/TLS"
  type        = bool
  default     = false
}

variable "ssl_certificate_id" {
  description = "SSL certificate ID"
  type        = string
  default     = null
}

variable "user_data" {
  description = "User data script for ECS instances"
  type        = string
  default     = ""
}