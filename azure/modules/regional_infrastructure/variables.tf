variable "region_prefix" {
  type        = string
  default     = "undefined"
  description = "Unique prefix for region"
}

variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "instance_type" {
  type = string
}

variable "disk_size" {
  type    = number
  default = 80
}

variable "delete_on_termination" {
  default     = false
  description = "Defines whether or not to delete data disks on termination. Useful when using scripts with CI. Must be lowercase"
  type        = bool
}

variable "validator_name" {
  default     = "Polkadot Failover validator"
  description = "A moniker of the validator"
}

variable "instance_count" {
  type    = number
  default = 1
}

variable "instance_count_primary" {
  type    = number
  default = 1
}

variable "instance_count_secondary" {
  type    = number
  default = 1
}

variable "instance_count_tertiary" {
  type    = number
  default = 1
}

variable "cpu_limit" {
  default     = "0.75"
  description = "CPU limit in CPUs number that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "ram_limit" {
  default     = "3.5"
  description = "RAM limit in GB that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "total_instance_count" {
  default = 3
}

variable "region" {
  type = string
}

variable "chain" {
  default     = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "vnet_cidr" {
  type = string
}

variable "subnet_cidr" {
  type = string
}

variable "subnet_cidrs" {
  type = list(string)
}

variable "rg" {
  type = string
}

variable "subscription" {
  type = string
}

variable "expose_ssh" {
  default = "False"
}

variable "ssh_key_content" {
  type = string
}

variable "ssh_user" {
  type = string
}

variable "admin_user" {
  type = string
}

variable "sa_type" {
  default     = "Standard_LRS"
  description = "Storage account type"
}

variable "tenant" {
}

variable "action_group_id" {
  type = string
}

variable "key_vault_name" {
  type = string
}

variable "wait_vmss" {
  description = "Should we wait until vmss is being ready. Required set az console utility"
  type        = bool
  default     = false
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
}