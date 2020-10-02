resource "azurerm_virtual_network_peering" "polkadot" {
  count                        = 6
  name                         = "peering-${count.index}"
  resource_group_name          = var.azure_rg
  virtual_network_name         = element(list(module.primary_region.vnet.name, module.secondary_region.vnet.name, module.tertiary_region.vnet.name), element(list(0, 0, 1, 1, 2, 2), count.index))
  remote_virtual_network_id    = element(list(module.primary_region.vnet.id, module.secondary_region.vnet.id, module.tertiary_region.vnet.id), element(list(1, 2, 0, 2, 0, 1), count.index))
  allow_virtual_network_access = true
  allow_forwarded_traffic      = true

  # `allow_gateway_transit` must be set to false for vnet Global Peering
  allow_gateway_transit = false
}
