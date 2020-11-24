variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "lbs" {
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

variable "vpc_id" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t3.small"
}

variable "key_name" {
  type    = string
  default = ""
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
