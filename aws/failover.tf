resource "polkadot_failover" "polkadot" {
  provider         = polkadot
  locations        = var.aws_regions
  instances        = var.instance_count
  prefix           = var.prefix
  metric_name      = var.validator_metric
  metric_namespace = var.prefix
  failover_mode    = var.failover_mode
}
