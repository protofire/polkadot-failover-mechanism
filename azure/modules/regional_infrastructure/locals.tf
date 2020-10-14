locals {

  lb_hc = {
    combo = ["Http", "8080", "/verify/checks"]
  }

  lb_port = {
    http     = ["8500", "Tcp", "8500"]
    dns      = ["8600", "Tcp", "8600"]
    rpc      = ["8300", "Tcp", "8300"]
    lan      = ["8301", "Tcp", "8301"]
    wan      = ["8302", "Tcp", "8302"]
    polkadot = ["30333", "Tcp", "30333"]
  }

  metric_namespaces = {
    health = {
      namespace = "${var.prefix}/health"
      //TODO: Value or value
      metric = "value"
    }
    health_checks = {
      namespace = "${var.prefix}/consul_health_checks"
      metric    = "critical"
    }
    disk = {
      namespace = "${var.prefix}/disk"
      metric    = "used_percent"
    }
  }

  monitoring_command_parameters = [
    var.subscription,
    var.rg,
    azurerm_linux_virtual_machine_scale_set.polkadot.name,
    "${local.metric_namespaces.health.namespace}:${local.metric_namespaces.health.metric}",
    "${local.metric_namespaces.health_checks.namespace}:${local.metric_namespaces.health_checks.metric}",
    "${local.metric_namespaces.disk.namespace}:${local.metric_namespaces.disk.metric}",
  ]
}
