output "prometheus_target" {
  value = "http://${aws_eip.prometheus.public_dns}:${var.prometheus_port}${var.metrics_path}"
}
