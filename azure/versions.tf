terraform {
  required_version = ">= 0.13"

  backend "azurerm" {
    version             = "~> 2.3"
    resource_group_name = var.azure_rg
  }

  required_providers {
    polkadot = {
      versions = ["0.1"]
      source   = "polkadot-failover-mechanism/azure/polkadot"
    }
  }

}
