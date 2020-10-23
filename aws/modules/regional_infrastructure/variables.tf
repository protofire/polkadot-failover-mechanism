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

variable "validator_keys" {
  type = map(object({
    seed = string
    key  = string
    type = string
  }))
}

variable "instance_count" {
  default = 1
}

variable "cpu_limit" {
  default     = "1.5"
  description = "CPU limit in CPUs number that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "ram_limit" {
  default     = "3.5"
  description = "RAM limit in GB that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "total_instance_count" {
  default = 3
}

variable "key_name" {
}

variable "key_content" {
  default = ""
}

variable "regions" {
  default = ["us-east-1", "us-east-2", "eu-west-1"]
}

variable "asg_role" {
}

variable "chain" {
  default     = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "vpc" {
}

variable "subnet" {
}

variable "lb" {
}

variable "lbs" {
}

variable "cidrs" {
}

variable "health_check_interval" {
}

variable "health_check_healthy_threshold" {
}

variable "health_check_unhealthy_threshold" {
}

variable "expose_ssh" {
  default = "false"
}

variable "node_key" {
  description = "A unique ed25519 key that identifies the node"
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
}
