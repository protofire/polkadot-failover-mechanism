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
  type = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "public_subnet_cidrs" {
  description = "Subnet CIDR for each region, must be different for VPC peering to work"
  default = ["10.0.0.0/24","10.1.0.0/24","10.2.0.0/24"]
}

variable "instance_type" {
  default = "n1-standard-1"
}

variable "cpu_limit" {
  default = "0.75"
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

variable "instance_count" {
  default = [1, 1, 1]
  description = "A number of instances to run in each region. Odd number of instances in total is a must have for proper work"
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
  default = "true"
}

variable "node_key" {
  description = "A unique ed25519 key that identifies the node"
}

variable "admin_email" {
  description = "An Admin email to send alerts to"
}

variable "gcp_ssh_user" {
  default = ""
}

variable "gcp_ssh_pub_key" {
  default = ""
}