data "google_compute_image" "centos" {
  family  = "centos-7"
  project = "centos-cloud"
}

resource "google_compute_address" "prometheus" {
  name         = "prometheus"
  address_type = "EXTERNAL"
}

resource "google_compute_instance" "prometheus" {
  project      = var.gcp_project != "" ? var.gcp_project : null
  name         = "${var.prefix}-polkadot-prometheus"
  machine_type = var.instance_type
  zone         = data.google_compute_zones.available.names[0]
  tags         = [var.prefix]

  labels = {
    prefix = var.prefix
  }

  description    = "Polkadot prometheus node"
  can_ip_forward = false

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }

  boot_disk {
    auto_delete = true
    initialize_params {
      image = data.google_compute_image.centos.self_link
    }
  }

  network_interface {
    subnetwork = var.subnets[0]
    access_config {
      nat_ip = google_compute_address.prometheus.address
    }
  }

  metadata_startup_script = templatefile("${path.module}/files/init.sh.tpl", {
    url_primary     = var.instance_count_primary > 0 ? "http://${google_compute_forwarding_rule.primary.ip_address}:${var.prometheus_port}${var.metrics_path}" : "",
    url_secondary   = var.instance_count_secondary > 0 ? "http://${google_compute_forwarding_rule.secondary.ip_address}:${var.prometheus_port}${var.metrics_path}" : "",
    url_tertiary    = var.instance_count_tertiary > 0 ? "http://${google_compute_forwarding_rule.tertiary.ip_address}:${var.prometheus_port}${var.metrics_path}" : "",
    prometheus_port = var.prometheus_port
  })

  metadata = {
    prefix   = var.prefix
    ssh-keys = var.expose_ssh && var.gcp_ssh_user != "" && var.gcp_ssh_pub_key != "" ? "${var.gcp_ssh_user}:${var.gcp_ssh_pub_key}" : null
  }

  service_account {
    email = var.sa_email
    scopes = [
      "compute-ro",
      "monitoring-write",
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }

}

data "google_compute_zones" "available" {
  project = var.gcp_project != "" ? var.gcp_project : null
  region  = var.gcp_regions[0]
}
