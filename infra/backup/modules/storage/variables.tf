# Variables for Storage Module

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

variable "private_subnet_ids" {
  description = "List of private subnet IDs"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs"
  type        = list(string)
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
}

variable "obs_buckets" {
  description = "Map of OBS bucket configurations"
  type = map(object({
    name          = string
    storage_class = string
    versioning    = bool
    encryption    = bool
    lifecycle_rules = optional(list(object({
      name                    = string
      prefix                  = string
      enabled                 = bool
      expiration_days         = optional(number)
      transition_days         = optional(number)
      transition_storage_class = optional(string)
    })), [])
  }))
  default = {}
}

variable "obs_encryption_enabled" {
  description = "Enable OBS bucket encryption"
  type        = bool
  default     = true
}

variable "obs_versioning_enabled" {
  description = "Enable OBS bucket versioning"
  type        = bool
  default     = true
}

variable "obs_lifecycle_rules" {
  description = "Default lifecycle rules for OBS buckets"
  type = list(object({
    name                    = string
    prefix                  = string
    enabled                 = bool
    expiration_days         = optional(number)
    transition_days         = optional(number)
    transition_storage_class = optional(string)
  }))
  default = []
}

variable "enable_database" {
  description = "Enable database deployment"
  type        = bool
  default     = false
}

variable "database_engine" {
  description = "Database engine type"
  type        = string
  default     = "postgresql"
}

variable "database_version" {
  description = "Database engine version"
  type        = string
  default     = "13"
}

variable "database_instance_class" {
  description = "Database instance class"
  type        = string
  default     = "rds.pg.n1.large.2"
}

variable "database_storage" {
  description = "Database storage size in GB"
  type        = number
  default     = 100
}

variable "database_backup_retention" {
  description = "Database backup retention in days"
  type        = number
  default     = 7
}

variable "enable_redis" {
  description = "Enable Redis deployment"
  type        = bool
  default     = false
}

variable "redis_instance_class" {
  description = "Redis instance class"
  type        = string
  default     = "redis.ha.xu1.large.r2.2"
}

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "5.0"
}

variable "enable_kms" {
  description = "Enable KMS for encryption"
  type        = bool
  default     = true
}

variable "kms_key_alias" {
  description = "KMS key alias"
  type        = string
  default     = "alias/pulse-expends"
}