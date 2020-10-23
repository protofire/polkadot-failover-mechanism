variable "azure_client" {
  description = "The Client ID which should be used. This can also be sourced from the ARM_CLIENT_ID Environment Variable."
  default     = null
  type        = string
}

variable "azure_client_secret" {
  description = "The Client Secret which should be used. This can also be sourced from the ARM_CLIENT_SECRET Environment Variable."
  default     = null
  type        = string
}

variable "azure_subscription" {
  description = "The Subscription ID which should be used. This can also be sourced from the ARM_SUBSCRIPTION_ID Environment Variable."
  default     = null
  type        = string
}

variable "azure_tenant" {
  description = "The Tenant ID which should be used. This can also be sourced from the ARM_TENANT_ID Environment Variable."
  default     = null
  type        = string
}


variable "use_msi" {
  default = true
  type    = bool
}

variable "azure_regions" {
  type        = list(string)
  default     = ["Central US", "East US", "West US"]
  description = "Should be an array consisting of exactly three elements"
  validation {
    condition     = length(var.azure_regions) == 3
    error_message = "Number of azure_client elements must be 3."
  }
}

variable "azure_rg" {
  type        = string
  description = "Resource group to create infrastructure at."
}

variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "public_vnet_cidrs" {
  type        = list(string)
  description = "VNet CIDR for each region, must be different for VPC peering to work"
  default     = ["10.0.0.0/16", "10.1.0.0/16", "10.2.0.0/16"]
  validation {
    condition     = length(var.public_vnet_cidrs) == 3
    error_message = "Number of public_vnet_cidrs elements must be 3."
  }
}

variable "public_subnet_cidrs" {
  type        = list(string)
  description = "Subnet CIDR for each region, must be different for VPC peering to work"
  default     = ["10.0.0.0/24", "10.1.0.0/24", "10.2.0.0/24"]
  validation {
    condition     = length(var.public_subnet_cidrs) == 3
    error_message = "Number of public_subnet_cidrs elements must be 3."
  }
}

variable "instance_type" {
  type    = string
  default = "Standard_D1_v2"
}

variable "cpu_limit" {
  type        = string
  default     = "0.75"
  description = "CPU limit in CPUs number that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "ram_limit" {
  type        = string
  default     = "3.5"
  description = "RAM limit in GB that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "disk_size" {
  type    = number
  default = 80
}

variable "delete_on_termination" {
  description = "Defines whether or not to delete data disks on termination. Useful when using scripts with CI"
  type        = bool
  default     = false
}

variable "validator_name" {
  type        = string
  default     = "Polkadot Failover validator"
  description = "A moniker of the validator"
}

variable "instance_count" {
  type        = list(number)
  default     = [1, 1, 1]
  description = "A number of instances to run in each region. Odd number of instances in total is a must have for proper work"
  validation {
    condition     = sum(var.instance_count) % 2 == 1 && length(var.instance_count) == 3
    error_message = "The sum of instance_count elements must be odd. Number of instance_count elements must be 3."
  }
}

variable "chain" {
  type        = string
  default     = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "validator_keys" {
  type = map(object({
    seed = string
    key  = string
    type = string
  }))
}

variable "ssh_key_content" {
  type = string
}

variable "ssh_user" {
  type    = string
  default = "polkadot"
}

variable "admin_user" {
  type    = string
  default = "polkadot"
}

variable "expose_ssh" {
  type    = string
  default = "False"
}

variable "sa_type" {
  type        = string
  default     = "Standard_LRS"
  description = "Storage account type"
}

variable "node_key" {
  type        = string
  description = "A unique ed25519 key that identifies the node"
}

variable "admin_email" {
  type        = string
  description = "An Admin email to send alerts to"
}

variable "wait_vmss" {
  description = "Should we wait until vmss is being ready. Required set az console utility"
  type        = bool
  default     = false
}

variable "failover_mode" {
  description = "Failover mode. Either 'single' or 'distributed'"
  type        = string
  default     = "distributed"
  validation {
    condition     = var.failover_mode == "single" || var.failover_mode == "distributed"
    error_message = "The failover_mode must be one of 'single', 'distributed'."
  }
}

variable "validate_metric" {
  description = "Name of telegraf validate metric"
  type        = string
  default     = "value"
}

variable "skip_provider_registration" {
  type    = bool
  default = true
}

variable "vault_soft_delete_enabled" {
  description = "Enable soft delete for key vault. Note, if you use a service principal it should has permissions in scope /subscription/`subscription_id`"
  type        = bool
  default     = false
}

variable "delete_vms_with_api_in_single_mode" {
  description = "Delete vms in single mode with API call preserving current active validator"
  type        = bool
  default     = false
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
  default     = "parity/polkadot:master"
}
