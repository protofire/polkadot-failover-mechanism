variable "aws_profiles" {
  type    = list(string)
  default = ["default"]
}

variable "aws_regions" {
  type        = list(string)
  default     = ["us-east-1", "us-east-2", "us-west-1"]
  description = "Should be an array consisting of exactly three elements"
}

variable "aws_access_keys" {
  type    = list(string)
  default = []
}

variable "aws_secret_keys" {
  type    = list(string)
  default = []
}

variable "prefix" {
  type = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "vpc_cidrs" {
  description = "VPC CIDR for each region, must be different for VPC peering to work"
  default = ["10.0.0.0/16","10.1.0.0/16","10.2.0.0/16"]
}

variable "public_subnet_cidrs" {
  description = "Subnet CIDR for each region, must be different for VPC peering to work"
  default = ["10.0.0.0/24","10.1.0.0/24","10.2.0.0/24"]
}

variable "instance_type" {
  default = "t3.medium"
}

variable "cpu_limit" {
  default = "1.5"
  description = "CPU limit in CPUs number that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "ram_limit" {
  default = "3.5"
  description = "RAM limit in GB that Polkadot node can use. Should never be greater than chosen instance type has."
}

variable "disk_size" {
  default = 80
}

variable "delete_on_termination" {
  default = "false"
  description = "Defines whether or not to delete data disks on termination. Useful when using scripts with CI"
}

variable "validator_name" {
  default = "Polkadot Failover validator"
  description = "A moniker of the validator"
}

variable "node_key" {
  description = "A unique ed25519 key that identifies the node"
}

variable "instance_count" {
  default = [1, 1, 1]
  description = "A number of instances to run in each region. Odd number of instances in total is a must have for proper work"
}

variable "key_name" {
}

variable "key_content" {
  default = ""
}

variable "chain" {
  default = "kusama"
  description = "A name of the chain to run Polkadot node at"
}

variable "health_check_interval" {
  default = 10
}

variable "health_check_healthy_threshold" {
  default = 3
}

variable "health_check_unhealthy_threshold" {
  default = 3
}

variable "validator_keys" {
  type = map(object({
    seed = string
    key = string
    type = string
  }))
}

variable "expose_ssh" {
  default = false
}
