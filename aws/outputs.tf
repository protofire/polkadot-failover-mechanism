output "lbs" {
  value = [module.primary_network.lb.arn, module.secondary_network.lb.arn, module.tertiary_network.lb.arn]
}

output "prometheus_target" {
  value = var.expose_prometheus ? module.prometheus[0].prometheus_target : ""
}
