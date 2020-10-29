variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "instance_type" {
  type    = string
  default = "t3.medium"
}

variable "disk_size" {
  type    = number
  default = 80
}

variable "delete_on_termination" {
  type        = bool
  default     = false
  description = "Defines whether or not to delete data disks on termination. Useful when using scripts with CI. Must be lowercase"
}

variable "validator_name" {
  type        = string
  default     = "Polkadot Failover validator"
  description = "A moniker of the validator"
}

variable "instance_count" {
  type    = number
  default = 1
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

variable "total_instance_count" {
  type    = number
  default = 3
}

variable "region" {
  type = string
}

variable "chain" {
  type        = string
  default     = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "subnet" {
  type = string
}

variable "health_check" {
  type = string
}

variable "expose_ssh" {
  type    = bool
  default = false
}

variable "sa_email" {
  type = string
}

variable "gcp_project" {
  type = string
}

variable "gcp_ssh_user" {
  type = string
}

variable "gcp_ssh_pub_key" {
  type = string
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
}

variable "metrics_namespace" {
  type    = string
  default = "polkadot"
}
