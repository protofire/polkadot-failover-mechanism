resource "google_compute_firewall" "prometheus" {
  project = var.gcp_project != "" ? var.gcp_project : null

  name        = "${var.prefix}-polkadot-prometheus"
  description = "Prometheus access"
  network     = var.network

  allow {
    ports    = [var.prometheus_port]
    protocol = "tcp"
  }

  priority      = 1000
  source_ranges = ["0.0.0.0/0"]
}
