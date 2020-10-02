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
  expose_ssh = "true"

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  regions = var.aws_regions
  cidrs   = var.vpc_cidrs

  instance_count       = var.instance_count[0]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]

  health_check_interval            = var.health_check_interval
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  providers = {
    aws = aws.primary
  }
}

module "secondary_region" {
  source = "./modules/regional_infrastructure"

  prefix        = var.prefix
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

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  asg_role   = aws_iam_instance_profile.monitoring.name
  expose_ssh = "true"

  regions = var.aws_regions
  cidrs   = var.vpc_cidrs

  instance_count       = var.instance_count[1]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]

  health_check_interval            = var.health_check_interval
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  providers = {
    aws = aws.secondary
  }
}

module "tertiary_region" {
  source = "./modules/regional_infrastructure"

  prefix        = var.prefix
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

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  asg_role   = aws_iam_instance_profile.monitoring.name
  expose_ssh = "true"

  regions = var.aws_regions
  cidrs   = var.vpc_cidrs

  instance_count       = var.instance_count[2]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]

  health_check_interval            = var.health_check_interval
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  providers = {
    aws = aws.tertiary
  }
}
