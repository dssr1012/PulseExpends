# Variables for Security Module

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

variable "project_id" {
  description = "Huawei Cloud project ID"
  type        = string
}

variable "iam_users" {
  description = "List of IAM users to create"
  type = list(object({
    name        = string
    enabled     = bool
    password    = string
    description = string
    role        = string
  }))
  default = []
}

variable "iam_groups" {
  description = "List of IAM groups to create"
  type = list(object({
    name        = string
    description = string
    members     = list(string)
  }))
  default = []
}

variable "iam_policies" {
  description = "List of IAM policies to create and attach"
  type = list(object({
    name            = string
    description     = string
    policy_document = string
    attached_groups = list(string)
    attached_users  = list(string)
  }))
  default = []
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

variable "enable_waf" {
  description = "Enable WAF protection"
  type        = bool
  default     = false
}

variable "waf_policy_name" {
  description = "WAF policy name"
  type        = string
  default     = "pulse-expends-waf-policy"
}

variable "enable_antiddos" {
  description = "Enable Anti-DDoS protection"
  type        = bool
  default     = true
}

variable "eip_id" {
  description = "EIP ID for Anti-DDoS protection"
  type        = string
  default     = null
}