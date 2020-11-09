locals {
  key_vault_name    = "${var.prefix}-polkadot-vault"
  metrics_namespace = "${var.prefix}/validator"

  metric_namespaces = {
    health = {
      namespace = "${var.prefix}/health"
      metric    = "value"
    }
    consul_health = {
      namespace = "${var.prefix}/consul_health_checks"
      metric    = "critical"
    }
    disk = {
      namespace = "${var.prefix}/disk"
      metric    = "used_percent"
    }
    validator = {
      namespace = "${var.prefix}/validator"
      metric    = "value"
    }
  }

  region_names = ["primary", "secondary", "tertiary"]
  scale_set_ids = [
    module.primary_region.scale_set_id,
    module.secondary_region.scale_set_id,
    module.tertiary_region.scale_set_id,
  ]
  scale_set_names = [
    module.primary_region.scale_set_name,
    module.secondary_region.scale_set_name,
    module.tertiary_region.scale_set_name,
  ]

}
