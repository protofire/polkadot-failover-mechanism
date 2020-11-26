output "prometheus_target" {
  value = var.expose_prometheus ? module.prometheus[0].prometheus_target : ""
}
