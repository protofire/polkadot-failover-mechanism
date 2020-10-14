resource "azurerm_network_security_group" "polkadot" {
  name                = "${var.prefix}-sg-${var.region_prefix}"
  location            = var.region
  resource_group_name = var.rg

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_network_security_rule" "outbound" {
  name                        = "out"
  priority                    = 100
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = var.rg
  network_security_group_name = azurerm_network_security_group.polkadot.name
}

resource "azurerm_network_security_rule" "ssh" {
  count                       = var.expose_ssh == "true" ? 1 : 0
  name                        = "ssh"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = var.rg
  network_security_group_name = azurerm_network_security_group.polkadot.name
}

resource "azurerm_network_security_rule" "p2pIn" {
  name                        = "p2p"
  priority                    = 101
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "30333"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = var.rg
  network_security_group_name = azurerm_network_security_group.polkadot.name
}

resource "azurerm_network_security_rule" "consul-0" {
  name                        = "consul-0"
  priority                    = 102
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_ranges     = ["8300", "8301", "8302", "8500", "8600"]
  source_address_prefix       = var.subnet_cidrs[0]
  destination_address_prefix  = "*"
  resource_group_name         = var.rg
  network_security_group_name = azurerm_network_security_group.polkadot.name
}

resource "azurerm_network_security_rule" "consul-1" {
  name                        = "consul-1"
  priority                    = 103
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_ranges     = ["8300", "8301", "8302", "8500", "8600"]
  source_address_prefix       = var.subnet_cidrs[1]
  destination_address_prefix  = "*"
  resource_group_name         = var.rg
  network_security_group_name = azurerm_network_security_group.polkadot.name
}

resource "azurerm_network_security_rule" "consul-2" {
  name                        = "consul-2"
  priority                    = 104
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_ranges     = ["8300", "8301", "8302", "8500", "8600"]
  source_address_prefix       = var.subnet_cidrs[2]
  destination_address_prefix  = "*"
  resource_group_name         = var.rg
  network_security_group_name = azurerm_network_security_group.polkadot.name
}

resource "azurerm_subnet_network_security_group_association" "polkadot" {
  subnet_id                 = azurerm_subnet.polkadot.id
  network_security_group_id = azurerm_network_security_group.polkadot.id
}
