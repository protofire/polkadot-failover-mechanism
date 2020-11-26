variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "subnets" {
  type = list(string)
}

variable "network" {
  type = string
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

variable "gcp_regions" {
  type = list(string)
}

variable "managed_groups" {
  type = list(any)
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

variable "instance_type" {
  type    = string
  default = "t3.small"
}

variable "expose_ssh" {
  type    = bool
  default = false
}

variable "prometheus_port" {
  type    = number
  default = 9273
}

variable "metrics_path" {
  type    = string
  default = "/metrics"
}
