variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
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
  type = string
}

variable "rg" {
  type = string
}

variable "admin_user" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "sa_type" {
  default     = "Standard_LRS"
  description = "Storage account type"
}

variable "region" {
  type = string
}

variable "ssh_key_content" {
  type = string
}

variable "ssh_user" {
  type = string
}

variable "subnet_cidrs" {
  type = list(string)
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
