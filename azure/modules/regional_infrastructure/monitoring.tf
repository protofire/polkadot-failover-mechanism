resource "azurerm_monitor_metric_alert" "health" {
  name                = "${var.prefix}-health-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance health is no OK."

  criteria {
    metric_namespace = "${var.prefix}/health"
    metric_name      = "Value"
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = var.action_group_id
  }

  depends_on = [null_resource.delay]
}

resource "azurerm_monitor_metric_alert" "health-status" {
  name                = "${var.prefix}-validator-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance health is no OK."

  criteria {
    metric_namespace = "${var.prefix}/consul_health_checks"
    metric_name      = "critical"
    aggregation      = "Minimum"
    operator         = "GreaterThan"
    threshold        = 0
  }

  action {
    action_group_id = var.action_group_id
  }

  depends_on = [null_resource.delay]
}

resource "azurerm_monitor_metric_alert" "disk" {
  name                = "${var.prefix}-disk-${var.region_prefix}"
  resource_group_name = var.rg
  scopes              = [azurerm_linux_virtual_machine_scale_set.polkadot.id]
  description         = "Action will be triggered when instance disk has no sufficient space."

  criteria {
    metric_namespace = "${var.prefix}/disk"
    metric_name      = "used_percent"
    aggregation      = "Maximum"
    operator         = "GreaterThan"
    threshold        = 90
  }

  action {
    action_group_id = var.action_group_id
  }

  depends_on = [null_resource.delay]
}

resource "null_resource" "delay" {
  provisioner "local-exec" {
    command = "sleep 550"
  }
  triggers = {
    before = azurerm_linux_virtual_machine_scale_set.polkadot.id
  }
}
