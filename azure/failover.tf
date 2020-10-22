data "polkadot_failover" "polkadot" {
  provider            = polkadot
  locations           = var.azure_regions
  instances           = var.instance_count
  prefix              = var.prefix
  metric_name         = var.validate_metric
  metric_namespace    = local.metrics_namespace
  failover_mode       = var.failover_mode
  resource_group_name = var.azure_rg
}
