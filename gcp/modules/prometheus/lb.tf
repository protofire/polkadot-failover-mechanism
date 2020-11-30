resource "google_compute_health_check" "healthcheck" {
  name                = "${var.prefix}-health-check"
  check_interval_sec  = 10
  timeout_sec         = 10
  healthy_threshold   = 2
  unhealthy_threshold = 10

  http_health_check {
    request_path = var.metrics_path
    port         = var.prometheus_port
  }
}

resource "google_compute_region_backend_service" "primary_backend_service" {
  name                  = "${var.prefix}-backend-service-primary"
  region                = var.gcp_regions[0]
  project               = var.gcp_project != "" ? var.gcp_project : null
  protocol              = "TCP"
  load_balancing_scheme = "INTERNAL"
  health_checks         = [google_compute_health_check.healthcheck.self_link]
  backend {
    group          = var.managed_groups[0].instance_group
    balancing_mode = ""
  }
}

resource "google_compute_region_backend_service" "secondary_backend_service" {
  name                  = "${var.prefix}-backend-service-secondary"
  region                = var.gcp_regions[1]
  project               = var.gcp_project != "" ? var.gcp_project : null
  protocol              = "TCP"
  load_balancing_scheme = "INTERNAL"
  health_checks         = [google_compute_health_check.healthcheck.self_link]
  backend {
    group = var.managed_groups[1].instance_group
  }
}

resource "google_compute_region_backend_service" "tertiary_backend_service" {
  name                  = "${var.prefix}-backend-service-tertiary"
  region                = var.gcp_regions[2]
  project               = var.gcp_project != "" ? var.gcp_project : null
  protocol              = "TCP"
  load_balancing_scheme = "INTERNAL"
  health_checks         = [google_compute_health_check.healthcheck.self_link]
  backend {
    group = var.managed_groups[2].instance_group
  }
}

resource "google_compute_forwarding_rule" "primary" {
  name                  = "${var.prefix}-forwarding-rule-primary"
  region                = var.gcp_regions[0]
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.primary_backend_service.id
  all_ports             = true
  allow_global_access   = true
  network               = var.network
  subnetwork            = var.subnets[0]
}

resource "google_compute_forwarding_rule" "secondary" {
  name                  = "${var.prefix}-forwarding-rule-secondary"
  region                = var.gcp_regions[1]
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.secondary_backend_service.id
  all_ports             = true
  allow_global_access   = true
  network               = var.network
  subnetwork            = var.subnets[1]
}

resource "google_compute_forwarding_rule" "tertiary" {
  name                  = "${var.prefix}-forwarding-rule-tertiary"
  region                = var.gcp_regions[2]
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.tertiary_backend_service.id
  all_ports             = true
  allow_global_access   = true
  network               = var.network
  subnetwork            = var.subnets[2]
}
