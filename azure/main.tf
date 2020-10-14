module "primary_region" {

  source = "./modules/regional_infrastructure"

  prefix        = var.prefix
  region_prefix = "primary"
  chain         = var.chain

  instance_count       = var.instance_count[0]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]
  instance_type        = var.instance_type

  region       = var.azure_regions[0]
  tenant       = var.azure_tenant
  rg           = var.azure_rg
  subscription = var.azure_subscription

  vnet_cidr    = var.public_vnet_cidrs[0]
  subnet_cidr  = var.public_subnet_cidrs[0]
  subnet_cidrs = var.public_subnet_cidrs

  expose_ssh      = var.expose_ssh
  ssh_key_content = var.ssh_key_content
  ssh_user        = var.ssh_user
  admin_user      = var.ssh_user

  disk_size             = var.disk_size
  sa_type               = var.sa_type
  delete_on_termination = var.delete_on_termination

  validator_name = var.validator_name

  key_vault_name = local.key_vault_name

  wait_vmss = var.wait_vmss

  action_group_id = azurerm_monitor_action_group.main.id

}

module "secondary_region" {
  source = "./modules/regional_infrastructure"

  region_prefix = "secondary"
  prefix        = var.prefix
  chain         = var.chain

  instance_count       = var.instance_count[1]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]
  instance_type        = var.instance_type

  region       = var.azure_regions[1]
  tenant       = var.azure_tenant
  rg           = var.azure_rg
  subscription = var.azure_subscription

  vnet_cidr    = var.public_vnet_cidrs[1]
  subnet_cidr  = var.public_subnet_cidrs[1]
  subnet_cidrs = var.public_subnet_cidrs

  expose_ssh      = var.expose_ssh
  ssh_key_content = var.ssh_key_content
  ssh_user        = var.ssh_user
  admin_user      = var.ssh_user

  disk_size             = var.disk_size
  sa_type               = var.sa_type
  delete_on_termination = var.delete_on_termination

  validator_name = var.validator_name

  key_vault_name = local.key_vault_name

  wait_vmss = var.wait_vmss

  action_group_id = azurerm_monitor_action_group.main.id

}

module "tertiary_region" {
  source = "./modules/regional_infrastructure"

  region_prefix = "tertiary"
  prefix        = var.prefix
  chain         = var.chain

  instance_count       = var.instance_count[2]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]
  instance_type        = var.instance_type

  region       = var.azure_regions[2]
  tenant       = var.azure_tenant
  rg           = var.azure_rg
  subscription = var.azure_subscription

  vnet_cidr    = var.public_vnet_cidrs[2]
  subnet_cidr  = var.public_subnet_cidrs[2]
  subnet_cidrs = var.public_subnet_cidrs

  expose_ssh      = var.expose_ssh
  ssh_key_content = var.ssh_key_content
  ssh_user        = var.ssh_user
  admin_user      = var.ssh_user

  disk_size             = var.disk_size
  sa_type               = var.sa_type
  delete_on_termination = var.delete_on_termination

  validator_name = var.validator_name

  key_vault_name = local.key_vault_name

  wait_vmss = var.wait_vmss

  action_group_id = azurerm_monitor_action_group.main.id

}
