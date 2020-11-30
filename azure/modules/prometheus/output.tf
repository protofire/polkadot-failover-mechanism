output "prometheus_target" {
  value = "http://${azurerm_public_ip.prometheus.ip_address}:${var.prometheus_port}${var.metrics_path}"
}
