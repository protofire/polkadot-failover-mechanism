data "azurerm_client_config" "current" {}

data "http" "my_public_ip" {
  url = "https://ifconfig.me"
}

resource "azurerm_key_vault" "polkadot" {
  name                       = local.key_vault_name
  location                   = var.azure_regions[0]
  resource_group_name        = var.azure_rg
  enabled_for_deployment     = true
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  soft_delete_enabled        = var.vault_soft_delete_enabled
  soft_delete_retention_days = 7
  purge_protection_enabled   = false
  enable_rbac_authorization  = false
  sku_name                   = "standard"

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = module.primary_region.principal_id

    storage_permissions = [
      "get",
      "list",
    ]

    key_permissions = [
      "get",
      "list",
    ]

    secret_permissions = [
      "get",
      "list",
    ]

  }

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    storage_permissions = [
      "get",
      "list",
      "delete",
      "purge",
      "recover",
      "restore",
    ]
    key_permissions = [
      "create",
      "get",
      "list",
      "delete",
      "purge",
      "recover",
    ]
    secret_permissions = [
      "set",
      "get",
      "list",
      "delete",
      "purge",
      "recover",
    ]
    certificate_permissions = [
      "get",
      "list",
      "delete",
      "purge",
      "recover",
    ]
  }

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = module.secondary_region.principal_id

    storage_permissions = [
      "get",
      "list",
    ]

    key_permissions = [
      "get",
      "list",
    ]

    secret_permissions = [
      "get",
      "list",
    ]
  }

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = module.tertiary_region.principal_id

    storage_permissions = [
      "get",
      "list",
    ]

    key_permissions = [
      "get",
      "list",
    ]

    secret_permissions = [
      "get",
      "list",
    ]
  }

  network_acls {
    ip_rules                   = ["${data.http.my_public_ip.body}/32"]
    virtual_network_subnet_ids = [module.primary_region.subnet, module.secondary_region.subnet, module.tertiary_region.subnet]
    default_action             = "Deny"
    bypass                     = "AzureServices"
  }

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "keys" {
  for_each = var.validator_keys

  name         = "polkadot-${var.prefix}-keys-${each.key}-key"
  value        = each.value.key
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "seeds" {
  for_each = var.validator_keys

  name         = "polkadot-${var.prefix}-keys-${each.key}-seed"
  value        = each.value.seed
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "types" {
  for_each = var.validator_keys

  name         = "polkadot-${var.prefix}-keys-${each.key}-type"
  value        = each.value.type
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "name" {
  name         = "polkadot-${var.prefix}-name"
  value        = var.validator_name
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "ram_limit" {
  name         = "polkadot-${var.prefix}-ramlimit"
  value        = var.ram_limit
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "cpu_limit" {
  name         = "polkadot-${var.prefix}-cpulimit"
  value        = var.cpu_limit
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_key_vault_secret" "node_key" {
  name         = "polkadot-${var.prefix}-nodekey"
  value        = var.node_key
  key_vault_id = azurerm_key_vault.polkadot.id

  tags = {
    prefix = var.prefix
  }
}
