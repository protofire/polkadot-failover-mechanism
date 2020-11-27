module "private_lb" {
  source = "../load_balancer"

  prefix        = var.prefix
  region_prefix = var.region_prefix

  location                               = var.region
  resource_group_name                    = var.rg
  type                                   = "private"
  frontend_subnet_id                     = azurerm_subnet.polkadot.id
  frontend_private_ip_address_allocation = "Static"
  frontend_private_ip_address            = local.lb_address.private_lb_address

  lb_hc = local.lb_hc

  lb_port = local.lb_port

  tags = {
    prefix = var.prefix
  }
}
