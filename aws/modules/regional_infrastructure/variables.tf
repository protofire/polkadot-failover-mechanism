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

variable "validator_keys" {
  type = map(object({
    seed = string
    key  = string
    type = string
  }))
}

variable "instance_count" {
  type    = number
  default = 1
}

variable "cpu_limit" {
  type        = string
  default     = "1.5"
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

variable "key_name" {
  type = string
}

variable "key_content" {
  type    = string
  default = ""
}

variable "regions" {
  type    = list(string)
  default = ["us-east-1", "us-east-2", "eu-west-1"]
}

variable "asg_role" {
  type = string
}

variable "chain" {
  type        = string
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
  type = list(any)
}

variable "cidrs" {
  type = list(string)
}

variable "health_check_interval" {
  type = number
}

variable "health_check_healthy_threshold" {
  type = number
}

variable "health_check_unhealthy_threshold" {
  type = number
}

variable "expose_ssh" {
  type    = bool
  default = false
}

variable "node_key" {
  type        = string
  description = "A unique ed25519 key that identifies the node"
}

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
}

variable "region" {
  type = string
}
