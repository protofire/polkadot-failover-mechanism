### THIS BLOCK ACCEPTS CONNECTIONS FROM PRIMARY REGION

resource "aws_vpc_peering_connection_accepter" "tertiary-primary" {
  provider                  = aws.tertiary
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-tertiary.id
  auto_accept               = true

  tags = {
    Side = "Accepter"
    prefix = var.prefix
  }
}

resource "aws_vpc_peering_connection_options" "tertiary-primary" {
  provider = aws.tertiary
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-tertiary.id

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  depends_on = [aws_vpc_peering_connection_accepter.tertiary-primary]
}

resource "aws_route" "tertiary-primary" {
  provider                  = aws.tertiary
  route_table_id            = module.tertiary_network.vpc.main_route_table_id
  destination_cidr_block    = var.vpc_cidrs[0]
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-tertiary.id

  timeouts {
    create = "5m"
  }
}

### THIS BLOCK ACCEPTS CONNECTIONS FROM SECONDARY REGION

resource "aws_vpc_peering_connection_accepter" "tertiary-secondary" {
  provider                  = aws.tertiary
  vpc_peering_connection_id = aws_vpc_peering_connection.secondary-tertiary.id
  auto_accept               = true

  tags = {
    Side = "Accepter"
    prefix = var.prefix
  }
}

resource "aws_vpc_peering_connection_options" "tertiary-secondary" {
  provider = aws.tertiary
  vpc_peering_connection_id = aws_vpc_peering_connection.secondary-tertiary.id

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  depends_on = [aws_vpc_peering_connection_accepter.tertiary-secondary]
}

resource "aws_route" "tertiary-secondary" {
  provider                  = aws.tertiary
  route_table_id            = module.tertiary_network.vpc.main_route_table_id
  destination_cidr_block    = var.vpc_cidrs[1]
  vpc_peering_connection_id = aws_vpc_peering_connection.secondary-tertiary.id

  timeouts {
    create = "5m"
  }
}
