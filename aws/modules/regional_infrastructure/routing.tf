# Create a gateway to provide access to the outside world
resource "aws_internet_gateway" "default" {
  vpc_id = var.vpc.id

  tags = {
    Name = "${var.prefix}-polkadot-validator"
    prefix = var.prefix
  }
}

# Grant the VPC internet access in its main route table
resource "aws_route" "internet_access" {
  route_table_id         = var.vpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.default.id
}