resource "google_compute_health_check" "autohealing" {
  name                = "${var.prefix}-autohealing-health-check"
  check_interval_sec  = var.health_check_interval
  timeout_sec         = 5
  healthy_threshold   = var.health_check_healthy_threshold 
  unhealthy_threshold = var.health_check_unhealthy_threshold

  http_health_check {
    request_path = "/verify/checks"
    port         = "8080"
  }
}
