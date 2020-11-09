resource "azurerm_monitor_action_group" "main" {
  name                = "${var.prefix}-ag"
  resource_group_name = var.azure_rg
  short_name          = var.prefix

  tags = {
    prefix = var.prefix
  }

  email_receiver {
    name                    = "Admin"
    email_address           = var.admin_email
    use_common_alert_schema = true
  }
}

resource "azurerm_monitor_metric_alert" "health" {
  count               = 3
  name                = "${var.prefix}-health-${local.region_names[count.index]}"
  scopes              = [local.scale_set_ids[count.index]]
  resource_group_name = var.azure_rg
  description         = "Action will be triggered when instance health is not OK."

  criteria {
    metric_namespace = local.metric_namespaces.health.namespace
    metric_name      = data.polkadot_metric_definition.health[count.index].metric_output_name
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  depends_on = [data.polkadot_metric_definition.health]
}

resource "azurerm_monitor_metric_alert" "consul_health" {
  count               = 3
  name                = "${var.prefix}-consul-health-${local.region_names[count.index]}"
  scopes              = [local.scale_set_ids[count.index]]
  resource_group_name = var.azure_rg
  description         = "Action will be triggered when instance consul health is no OK."

  criteria {
    metric_namespace = local.metric_namespaces.consul_health.namespace
    metric_name      = data.polkadot_metric_definition.consul_health[count.index].metric_output_name
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  depends_on = [data.polkadot_metric_definition.consul_health]

}

resource "azurerm_monitor_metric_alert" "disk" {
  count               = 3
  name                = "${var.prefix}-disk-${local.region_names[count.index]}"
  scopes              = [local.scale_set_ids[count.index]]
  resource_group_name = var.azure_rg
  description         = "Action will be triggered when instance disk has no sufficient space."

  criteria {
    metric_namespace = local.metric_namespaces.disk.namespace
    metric_name      = data.polkadot_metric_definition.disk[count.index].metric_output_name
    aggregation      = "Maximum"
    operator         = "GreaterThan"
    threshold        = 90
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  depends_on = [data.polkadot_metric_definition.disk]

}

data "polkadot_metric_definition" "disk" {
  provider            = polkadot
  count               = 3
  scale_sets          = [local.scale_set_names[count.index]]
  prefix              = var.prefix
  metric_namespace    = local.metric_namespaces.disk.namespace
  metric_name         = local.metric_namespaces.disk.metric
  resource_group_name = var.azure_rg
  depends_on = [
    module.primary_region,
    module.secondary_region,
    module.tertiary_region,
    azurerm_key_vault.polkadot,
    azurerm_key_vault_secret.cpu_limit,
    azurerm_key_vault_secret.keys,
    azurerm_key_vault_secret.node_key,
    azurerm_key_vault_secret.ram_limit,
    azurerm_key_vault_secret.seeds,
    azurerm_key_vault_secret.types,
    azurerm_key_vault_secret.name,
  ]
}

data "polkadot_metric_definition" "health" {
  provider            = polkadot
  count               = 3
  scale_sets          = [local.scale_set_names[count.index]]
  prefix              = var.prefix
  metric_namespace    = local.metric_namespaces.health.namespace
  metric_name         = local.metric_namespaces.health.metric
  resource_group_name = var.azure_rg
  depends_on = [
    module.primary_region,
    module.secondary_region,
    module.tertiary_region,
    azurerm_key_vault.polkadot,
    azurerm_key_vault_secret.cpu_limit,
    azurerm_key_vault_secret.keys,
    azurerm_key_vault_secret.node_key,
    azurerm_key_vault_secret.ram_limit,
    azurerm_key_vault_secret.seeds,
    azurerm_key_vault_secret.types,
    azurerm_key_vault_secret.name,
  ]
}

data "polkadot_metric_definition" "consul_health" {
  provider            = polkadot
  count               = 3
  scale_sets          = [local.scale_set_names[count.index]]
  prefix              = var.prefix
  metric_namespace    = local.metric_namespaces.consul_health.namespace
  metric_name         = local.metric_namespaces.consul_health.metric
  resource_group_name = var.azure_rg
  depends_on = [
    module.primary_region,
    module.secondary_region,
    module.tertiary_region,
    azurerm_key_vault.polkadot,
    azurerm_key_vault_secret.cpu_limit,
    azurerm_key_vault_secret.keys,
    azurerm_key_vault_secret.node_key,
    azurerm_key_vault_secret.ram_limit,
    azurerm_key_vault_secret.seeds,
    azurerm_key_vault_secret.types,
    azurerm_key_vault_secret.name,
  ]
}
