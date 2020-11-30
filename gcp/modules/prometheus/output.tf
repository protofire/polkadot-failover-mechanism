output "prometheus_target" {
  value = "http://${google_compute_address.prometheus.address}:${var.prometheus_port}${var.metrics_path}"
}
