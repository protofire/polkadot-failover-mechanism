resource "polkadot_failover" "polkadot" {
  provider         = polkadot
  locations        = var.gcp_regions
  instances        = var.instance_count
  prefix           = var.prefix
  metric_name      = local.validator_metric_name
  metric_namespace = var.metric_namespace
  failover_mode    = var.failover_mode
}
