resource "azurerm_monitor_metric_alert" "health" {
  name                = "${var.prefix}-health-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance health is no OK."

  count = azurerm_linux_virtual_machine_scale_set.polkadot.instances > 0 ? 1 : 0

  criteria {
    metric_namespace = local.metric_namespaces.health.namespace
    metric_name      = data.external.monitoring_namespaces[0].result[local.metric_namespaces.health.namespace]
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = var.action_group_id
  }

  lifecycle {
    ignore_changes = [
      criteria,
    ]
  }


  depends_on = [data.external.monitoring_namespaces]
}

resource "azurerm_monitor_metric_alert" "health-status" {
  name                = "${var.prefix}-validator-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance health is no OK."

  count = azurerm_linux_virtual_machine_scale_set.polkadot.instances > 0 ? 1 : 0

  criteria {
    metric_namespace = local.metric_namespaces.health_checks.namespace
    metric_name      = data.external.monitoring_namespaces[0].result[local.metric_namespaces.health_checks.namespace]
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = var.action_group_id
  }

  lifecycle {
    ignore_changes = [
      criteria,
    ]
  }

  depends_on = [data.external.monitoring_namespaces]

}

resource "azurerm_monitor_metric_alert" "disk" {
  name                = "${var.prefix}-disk-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance disk has no sufficient space."

  count = azurerm_linux_virtual_machine_scale_set.polkadot.instances > 0 ? 1 : 0

  criteria {
    metric_namespace = local.metric_namespaces.disk.namespace
    metric_name      = data.external.monitoring_namespaces[0].result[local.metric_namespaces.disk.namespace]
    aggregation      = "Maximum"
    operator         = "GreaterThan"
    threshold        = 90
  }

  action {
    action_group_id = var.action_group_id
  }

  lifecycle {
    ignore_changes = [
      criteria,
    ]
  }

  depends_on = [data.external.monitoring_namespaces]

}

data "external" "monitoring_namespaces" {

  depends_on = [azurerm_linux_virtual_machine_scale_set.polkadot]

  count = azurerm_linux_virtual_machine_scale_set.polkadot.instances > 0 ? 1 : 0

  program = flatten([
    "bash",
    "${path.module}/../../../init-helpers/azure/wait_monitoring_namespace.sh",
    flatten(local.monitoring_command_parameters)
  ])

}
