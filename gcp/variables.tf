variable "gcp_credentials" {
  type    = string
  default = ""
}

variable "gcp_regions" {
  type        = list(string)
  default     = ["us-east1", "us-east4", "us-west1"]
  description = "Should be an array consisting of exactly three elements"
}

variable "gcp_project" {
  type    = string
  default = ""
}

variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "public_subnet_cidrs" {
  type        = list(string)
  description = "Subnet CIDR for each region, must be different for VPC peering to work"
  default     = ["10.0.0.0/24", "10.1.0.0/24", "10.2.0.0/24"]
}

variable "instance_type" {
  type    = string
  default = "n1-standard-1"
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
  type        = bool
  default     = false
  description = "Defines whether or not to delete data disks on termination. Useful when using scripts with CI"
}

variable "validator_name" {
  type        = string
  default     = "Polkadot Failover validator"
  description = "A moniker of the validator"
}

variable "chain" {
  type        = string
  default     = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "health_check_interval" {
  type    = number
  default = 10
}

variable "health_check_healthy_threshold" {
  type    = number
  default = 3
}

variable "health_check_unhealthy_threshold" {
  type    = number
  default = 3
}

variable "validator_keys" {
  type = map(object({
    seed = string
    key  = string
    type = string
  }))
}

variable "expose_ssh" {
  type    = bool
  default = false
}

variable "node_key" {
  type        = string
  description = "A unique ed25519 key that identifies the node"
}

variable "admin_email" {
  type        = string
  description = "An Admin email to send alerts to"
}

variable "gcp_ssh_user" {
  type    = string
  default = ""
}

variable "gcp_ssh_pub_key" {
  type    = string
  default = ""
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

variable "failover_mode" {
  description = "Failover mode. Either 'single' or 'distributed'"
  type        = string
  default     = "distributed"
  validation {
    condition     = var.failover_mode == "single" || var.failover_mode == "distributed"
    error_message = "The failover_mode must be one of 'single', 'distributed'."
  }
}

variable "delete_vms_with_api_in_single_mode" {
  description = "Delete vms in single mode with API call preserving current active validator"
  type        = bool
  default     = true
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
  default     = "parity/polkadot:master-0.8.26-80e3a7e-5d52c096"
}

variable "metric_namespace" {
  type    = string
  default = "polkadot"
}

variable "metric_name" {
  description = "Name of telegraf validate metric"
  type        = string
  default     = "value"
}

variable "metric_family" {
  description = "Name of telegraf validate metrics family"
  type        = string
  default     = "validator"
}
