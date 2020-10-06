resource "azurerm_monitor_metric_alert" "health" {
  name                = "${var.prefix}-health-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance health is no OK."

  criteria {
    metric_namespace = local.metric_namespaces.health.namespace
    metric_name      = local.metric_namespaces.health.metric
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = var.action_group_id
  }

  depends_on = [null_resource.monitoring_namespaces]
}

resource "azurerm_monitor_metric_alert" "health-status" {
  name                = "${var.prefix}-validator-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance health is no OK."

  criteria {
    metric_namespace = local.metric_namespaces.health_checks.namespace
    metric_name      = local.metric_namespaces.health_checks.metric
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = var.action_group_id
  }

  depends_on = [null_resource.monitoring_namespaces]

}

resource "azurerm_monitor_metric_alert" "disk" {
  name                = "${var.prefix}-disk-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance disk has no sufficient space."

  criteria {
    metric_namespace = local.metric_namespaces.disk.namespace
    metric_name      = local.metric_namespaces.disk.metric
    aggregation      = "Maximum"
    operator         = "GreaterThan"
    threshold        = 90
  }

  action {
    action_group_id = var.action_group_id
  }

  depends_on = [null_resource.monitoring_namespaces]

}

resource "null_resource" "monitoring_namespaces" {

  depends_on = [azurerm_linux_virtual_machine_scale_set.polkadot]

  provisioner "local-exec" {
    command = "bash ${path.module}/../../../init-helpers/azure/wait_monitoring_namespace.sh ${join(" ", local.monitoring_command_parameters)}"
  }

}
