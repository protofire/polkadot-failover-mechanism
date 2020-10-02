provider "azurerm" {
  version = "~> 2.10"
  features {}
  client_id       = var.azure_client != "" ? var.azure_client : null
  subscription_id = var.azure_subscription != "" ? var.azure_subscription : null
  tenant_id       = var.azure_tenant != "" ? var.azure_tenant : null
}

provider "http" {
  version = "~> 1.2"
}
