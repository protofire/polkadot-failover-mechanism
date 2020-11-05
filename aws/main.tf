module "primary_network" {
  source = "./modules/regional_networks"

  prefix      = var.prefix
  vpc_cidr    = var.vpc_cidrs[0]
  subnet_cidr = var.public_subnet_cidrs[0]

  providers = {
    aws = aws.primary
  }
}

module "secondary_network" {
  source = "./modules/regional_networks"

  prefix      = var.prefix
  vpc_cidr    = var.vpc_cidrs[1]
  subnet_cidr = var.public_subnet_cidrs[1]

  providers = {
    aws = aws.secondary
  }
}

module "tertiary_network" {
  source = "./modules/regional_networks"

  prefix      = var.prefix
  vpc_cidr    = var.vpc_cidrs[2]
  subnet_cidr = var.public_subnet_cidrs[2]

  providers = {
    aws = aws.tertiary
  }
}


module "primary_region" {
  source = "./modules/regional_infrastructure"

  prefix        = var.prefix
  region_prefix = local.region_prefixes[0]
  vpc           = module.primary_network.vpc
  subnet        = module.primary_network.subnet
  lb            = module.primary_network.lb
  lbs           = [module.primary_network.lb, module.secondary_network.lb, module.tertiary_network.lb]
  instance_type = var.instance_type

  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  validator_name = var.validator_name
  validator_keys = var.validator_keys
  node_key       = var.node_key
  key_name       = var.key_name
  key_content    = var.key_content
  chain          = var.chain

  asg_role   = aws_iam_instance_profile.monitoring.name
  expose_ssh = var.expose_ssh

  region    = var.aws_regions[0]
  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  regions = var.aws_regions
  cidrs   = var.vpc_cidrs

  instance_count       = polkadot_failover.polkadot.primary_count
  total_instance_count = sum(polkadot_failover.polkadot.failover_instances)

  health_check_interval            = var.health_check_interval
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  providers = {
    aws = aws.primary
  }
  docker_image = var.docker_image
}

module "secondary_region" {
  source = "./modules/regional_infrastructure"

  prefix        = var.prefix
  region_prefix = local.region_prefixes[1]
  vpc           = module.secondary_network.vpc
  subnet        = module.secondary_network.subnet
  lb            = module.secondary_network.lb
  lbs           = [module.primary_network.lb, module.secondary_network.lb, module.tertiary_network.lb]
  instance_type = var.instance_type

  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  validator_name = var.validator_name
  validator_keys = var.validator_keys
  node_key       = var.node_key
  key_name       = var.key_name
  key_content    = var.key_content
  chain          = var.chain

  region    = var.aws_regions[1]
  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  asg_role   = aws_iam_instance_profile.monitoring.name
  expose_ssh = var.expose_ssh

  regions = var.aws_regions
  cidrs   = var.vpc_cidrs

  instance_count       = polkadot_failover.polkadot.secondary_count
  total_instance_count = sum(polkadot_failover.polkadot.failover_instances)

  health_check_interval            = var.health_check_interval
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  providers = {
    aws = aws.secondary
  }
  docker_image = var.docker_image
}

module "tertiary_region" {
  source = "./modules/regional_infrastructure"

  prefix        = var.prefix
  region_prefix = local.region_prefixes[2]
  vpc           = module.tertiary_network.vpc
  subnet        = module.tertiary_network.subnet
  lb            = module.tertiary_network.lb
  lbs           = [module.primary_network.lb, module.secondary_network.lb, module.tertiary_network.lb]
  instance_type = var.instance_type

  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  validator_name = var.validator_name
  validator_keys = var.validator_keys
  node_key       = var.node_key
  key_name       = var.key_name
  key_content    = var.key_content
  chain          = var.chain

  region    = var.aws_regions[2]
  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  asg_role   = aws_iam_instance_profile.monitoring.name
  expose_ssh = var.expose_ssh

  regions = var.aws_regions
  cidrs   = var.vpc_cidrs

  instance_count       = polkadot_failover.polkadot.tertiary_count
  total_instance_count = sum(polkadot_failover.polkadot.failover_instances)

  health_check_interval            = var.health_check_interval
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  providers = {
    aws = aws.tertiary
  }
  docker_image = var.docker_image
}
