output "vnet" {
  value = azurerm_virtual_network.polkadot
}

output "subnet_id" {
  value = azurerm_subnet.polkadot.id
}

output "principal_id" {
  value = azurerm_linux_virtual_machine_scale_set.polkadot.identity[0].principal_id
}

output "scale_set_id" {
  value = azurerm_linux_virtual_machine_scale_set.polkadot.id
}

output "scale_set_name" {
  value = azurerm_linux_virtual_machine_scale_set.polkadot.name
}

output "private_lb_id" {
  value = module.private_lb.azurerm_lb_id
}

output "private_lb_address" {
  value = local.lb_address.private_lb_address
}
