variable "aws_profiles" {
  type    = list(string)
  default = ["default"]
}

variable "aws_regions" {
  type        = list(string)
  default     = ["us-east-1", "us-east-2", "us-west-1"]
  description = "Should be an array consisting of exactly three elements"
  validation {
    condition     = length(var.aws_regions) == 3
    error_message = "Number of aws_regions elements must be 3."
  }
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
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "vpc_cidrs" {
  type        = list(string)
  description = "VPC CIDR for each region, must be different for VPC peering to work"
  default     = ["10.0.0.0/16", "10.1.0.0/16", "10.2.0.0/16"]
  validation {
    condition     = length(var.vpc_cidrs) == 3
    error_message = "Number of vpc_cidrs elements must be 3."
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
  default = "t3.medium"
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

variable "node_key" {
  type        = string
  description = "A unique ed25519 key that identifies the node"
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

variable "key_name" {
  type = string
}

variable "key_content" {
  type    = string
  default = ""
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

variable "docker_image" {
  description = "Polkadot docker image"
  type        = string
  default     = "parity/polkadot:master-0.8.26-80e3a7e-5d52c096"
}
