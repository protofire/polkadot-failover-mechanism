variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "instance_type" {
  default = "t3.medium"
}

variable "disk_size" {
  default = 80
}

variable "delete_on_termination" {
  default     = "false"
  description = "Defines whether or not to delete data disks on termination. Useful when using scripts with CI. Must be lowercase"
}

variable "validator_name" {
  default     = "Polkadot Failover validator"
  description = "A moniker of the validator"
}

variable "instance_count" {
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
}

variable "chain" {
  default     = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "subnet" {
}

variable "health_check" {
}

variable "expose_ssh" {
  default = "false"
}

variable "sa_email" {
}

variable "gcp_project" {
}

variable "gcp_ssh_user" {
}

variable "gcp_ssh_pub_key" {
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
}
