### THIS BLOCK CREATES CONNECTIONS TO SECONDARY REGION

resource "aws_vpc_peering_connection" "primary-secondary" {
  provider      = aws.primary
  peer_vpc_id   = module.secondary_network.vpc.id
  peer_region   = var.aws_regions[1]
  
  vpc_id        = module.primary_network.vpc.id

  auto_accept   = false

  tags = {
    Side = "Requester"
    prefix = var.prefix
  }
}

resource "aws_route" "primary-secondary" {
  provider                  = aws.primary
  route_table_id            = module.primary_network.vpc.main_route_table_id
  destination_cidr_block    = var.vpc_cidrs[1]
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-secondary.id
}

resource "aws_vpc_peering_connection_options" "primary-secondary" {
  provider = aws.primary
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-secondary.id

  requester {
    allow_remote_vpc_dns_resolution = true
  }

  depends_on = [aws_vpc_peering_connection_accepter.secondary-primary]
}

### THIS BLOCK CREATES CONNECTIONS TO TERTIARY REGION

resource "aws_vpc_peering_connection_options" "primary-tertiary" {
  provider = aws.primary
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-tertiary.id

  requester {
    allow_remote_vpc_dns_resolution = true
  }

  depends_on = [aws_vpc_peering_connection_accepter.tertiary-primary]
}

resource "aws_vpc_peering_connection" "primary-tertiary" {
  provider      = aws.primary
  peer_vpc_id   = module.tertiary_network.vpc.id
  peer_region   = var.aws_regions[2]
  
  vpc_id        = module.primary_network.vpc.id

  auto_accept   = false

  tags = {
    Side = "Requester"
    prefix = var.prefix
  }
}

resource "aws_route" "primary-tertiary" {
  provider                  = aws.primary
  route_table_id            = module.primary_network.vpc.main_route_table_id
  destination_cidr_block    = var.vpc_cidrs[2]
  vpc_peering_connection_id = aws_vpc_peering_connection.primary-tertiary.id
}