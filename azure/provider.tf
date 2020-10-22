provider "azurerm" {
  version                    = "~> 2.3"
  client_id                  = var.azure_client != "" ? var.azure_client : null
  client_secret              = var.azure_client_secret != "" ? var.azure_client_secret : null
  subscription_id            = var.azure_subscription != "" ? var.azure_subscription : null
  tenant_id                  = var.azure_tenant != "" ? var.azure_tenant : null
  use_msi                    = var.use_msi
  skip_provider_registration = var.skip_provider_registration

  features {
    key_vault {
      recover_soft_deleted_key_vaults = false
      purge_soft_delete_on_destroy    = var.vault_soft_delete_enabled
    }
  }

}

provider "http" {
  version = "~> 1.2"
}

provider "polkadot" {
  version                            = "~> 0.1"
  client_id                          = var.azure_client != "" ? var.azure_client : null
  client_secret                      = var.azure_client_secret != "" ? var.azure_client_secret : null
  subscription_id                    = var.azure_subscription != "" ? var.azure_subscription : null
  tenant_id                          = var.azure_tenant != "" ? var.azure_tenant : null
  use_msi                            = var.use_msi
  skip_provider_registration         = var.skip_provider_registration
  delete_vms_with_api_in_single_mode = var.delete_vms_with_api_in_single_mode
}
