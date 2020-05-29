data "google_compute_image" "centos" {
  family  = "centos-7"
  project = "centos-cloud"
}

resource "google_compute_instance_template" "instance_template" {
  name_prefix  = "${var.prefix}-"
  machine_type = var.instance_type
  region       = var.region
  description = "This template is used to create Polkadot failover nodes."

  tags = [var.prefix]

  labels = {
    prefix = var.prefix
    cluster-size = var.total_instance_count
  }

  instance_description = "Polkadot failover node"
  can_ip_forward       = false

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }

  // Create a new boot disk from an image
   disk {
    source_image = data.google_compute_image.centos.self_link
    auto_delete  = true
    boot         = true
    disk_size_gb = 20
  }

  disk {
    auto_delete  = var.delete_on_termination
    device_name  = "sdb"
    boot         = false
    disk_size_gb = var.disk_size
  }

  network_interface {
    subnetwork = var.subnet
    access_config {
    }
  }

  metadata_startup_script = templatefile("${path.module}/files/init.sh.tpl", { prefix = var.prefix, chain = var.chain, total_instance_count = var.total_instance_count })

  metadata = {
    shutdown-script = templatefile("${path.module}/files/shutdown.sh.tpl", {})
    prefix = var.prefix
  }

  service_account {
    email = var.sa_email
    scopes = ["compute-ro", "https://www.googleapis.com/auth/cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_instance_group_manager" "instance_group_manager" {
  name               = "${var.prefix}-instance-group-manager"
  version {
    instance_template  = google_compute_instance_template.instance_template.self_link
  }
  
  named_port {
    name = "polkadot"
    port = "30333"
  }

  named_port {
    name = "consul-rpc"
    port = "8300"
  }

  named_port {
    name = "consul-lan"
    port = "8301"
  }

  named_port {
    name = "consul-http"
    port = "8500"
  }

 named_port {
    name = "consul-dns"
    port = "8600"
  }

  base_instance_name = "${var.prefix}-polkadot-failover-instance"
  region             = var.region
  target_size        = var.instance_count

  update_policy {
    type                         = "PROACTIVE"
    instance_redistribution_type = "PROACTIVE"
    minimal_action               = "REPLACE"
    max_surge_fixed              = 0
    max_unavailable_fixed        = length(data.google_compute_zones.available.names)
    min_ready_sec                = 150
  }

  auto_healing_policies {
    health_check      = var.health_check 
    initial_delay_sec = 200
  }

}

data "google_compute_zones" "available" {
}