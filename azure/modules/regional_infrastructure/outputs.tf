output "vnet" {
  value = azurerm_virtual_network.polkadot
}

output "subnet" {
  value = azurerm_subnet.polkadot.id
}

output "principal_id" {
  value = azurerm_linux_virtual_machine_scale_set.polkadot.identity[0].principal_id
}
