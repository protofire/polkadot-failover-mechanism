variable "prefix" {
  type        = string
  description = "Unique prefix for cloud resources at Terraform"
}

variable "vpc_cidr" {
  type = string
}

variable "subnet_cidr" {
  type = string
}