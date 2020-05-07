### THIS BLOCK ACCEPTS CONNECTIONS FROM PRIMARY REGION

resource "aws_vpc_peering_connection_accepter" "secondary-primary" {
  provider                  = aws.secondary
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-secondary.id
  auto_accept               = true

  tags = {
    Side = "Accepter"
    prefix = var.prefix
  }
}

resource "aws_vpc_peering_connection_options" "secondary-primary" {
  provider = aws.secondary
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-secondary.id

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  depends_on = [aws_vpc_peering_connection_accepter.secondary-primary]
}

resource "aws_route" "secondary-primary" {
  provider                  = aws.secondary
  route_table_id            = module.secondary_network.vpc.main_route_table_id
  destination_cidr_block    = var.vpc_cidrs[0]
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-secondary.id

  timeouts {
    create = "5m"
  }
}

### THIS BLOCK CREATES CONNECTIONS TO TERTIARY REGION

resource "aws_vpc_peering_connection" "secondary-tertiary" {
  provider      = aws.secondary
  peer_vpc_id   = module.tertiary_network.vpc.id
  peer_region   = var.aws_regions[2]
  
  vpc_id        = module.secondary_network.vpc.id

  auto_accept   = false

  tags = {
    Side = "Requester"
    prefix = var.prefix
  }
}

resource "aws_route" "secondary-tertiary" {
  provider                  = aws.secondary
  route_table_id            = module.secondary_network.vpc.main_route_table_id
  destination_cidr_block    = var.vpc_cidrs[2]
  vpc_peering_connection_id = aws_vpc_peering_connection.secondary-tertiary.id

  timeouts {
    create = "5m"
  }
}

resource "aws_vpc_peering_connection_options" "secondary-tertiary" {
  provider = aws.secondary
  vpc_peering_connection_id = aws_vpc_peering_connection.secondary-tertiary.id

  requester {
    allow_remote_vpc_dns_resolution = true
  }

  depends_on = [aws_vpc_peering_connection_accepter.tertiary-secondary]
}
