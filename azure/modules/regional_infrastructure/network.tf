resource "azurerm_virtual_network" "polkadot" {
  name                = "${var.prefix}-network-${var.region_prefix}"
  address_space       = [var.vnet_cidr]
  location            = var.region
  resource_group_name = var.rg
}

resource "azurerm_subnet" "polkadot" {
  name                 = "${var.prefix}-subnet-${var.region_prefix}"
  resource_group_name  = var.rg
  virtual_network_name = azurerm_virtual_network.polkadot.name
  address_prefixes     = [var.subnet_cidr]
  service_endpoints    = ["Microsoft.KeyVault"]

  /*
  provisioner "local-exec" {
    command = "az keyvault network-rule add --name ${var.key_vault_name} --subnet ${azurerm_subnet.polkadot.id}"
  }
*/
}
