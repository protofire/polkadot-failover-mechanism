resource "google_compute_firewall" "health-check" {
  name        = "${var.prefix}-polkadot-validator-hc"
  description = "A security group health checks"
  network     = google_compute_network.vpc_network.self_link

  allow {
    ports           = ["8080"]
    protocol        = "tcp"
  }
 
  priority        = 1003
  source_ranges   = ["35.191.0.0/16", "130.211.0.0/22"]
}


resource "google_compute_firewall" "validator-node-internal" {
  name        = "${var.prefix}-polkadot-validator-internal"
  description = "Security for consul communications"
  network     = google_compute_network.vpc_network.self_link

  allow {
    ports           = ["8500","8600","8300","8301","8302"]
    protocol        = "tcp"
  }
 
  allow {
    ports           = ["8500","8600","8301","8302"]
    protocol        = "udp"
  }
  
  priority        = 1001
  source_tags     = [var.prefix]
}

resource "google_compute_firewall" "validator-node-external" {
  name        = "${var.prefix}-polkadot-validator-external"
  description = "For blockchain node to be accessible from outside, also for SSH access if configured"
  network     = google_compute_network.vpc_network.self_link 

  dynamic "allow" {
    for_each = var.expose_ssh == "false" ? [] : [1]
    content {
        ports         = ["22"]
        protocol      = "tcp"
    }
  }

  allow {
    ports           = ["30333"]
    protocol        = "tcp"
  }

  allow {
    ports           = ["30333"]
    protocol        = "udp"
  }
 
  priority      = 1000
  source_ranges = ["0.0.0.0/0"] 
}
