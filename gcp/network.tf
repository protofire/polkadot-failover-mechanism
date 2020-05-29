resource "google_compute_network" "vpc_network" {
  provider = google.primary

  name                    = "${var.prefix}-polkadot-failover"
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"
}

resource "google_compute_subnetwork" "primary" {
  provider = google.primary

  name          = "${var.prefix}-subnetwork-primary"
  ip_cidr_range = var.public_subnet_cidrs[0]
  network       = google_compute_network.vpc_network.self_link
}

resource "google_compute_subnetwork" "secondary" {
  provider = google.secondary

  name          = "${var.prefix}-subnetwork-secondary"
  ip_cidr_range = var.public_subnet_cidrs[1]
  network       = google_compute_network.vpc_network.self_link
}

resource "google_compute_subnetwork" "tertiary" {
  provider = google.tertiary

  name          = "${var.prefix}-subnetwork-tertiary"
  ip_cidr_range = var.public_subnet_cidrs[2]
  network       = google_compute_network.vpc_network.self_link
}


