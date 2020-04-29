resource "aws_vpc" "vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name   = "${var.prefix}-polkadot-validator"
    prefix = var.prefix
  }
}

data "aws_availability_zones" "available" {
}

resource "aws_subnet" "default" {
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = var.subnet_cidr
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name   = "${var.prefix}-default-subnet"
    prefix = var.prefix
  }
}