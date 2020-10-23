module "primary_region" {
  source = "./modules/regional_infrastructure"

  sa_email = google_service_account.service_account.email

  prefix      = var.prefix
  subnet      = google_compute_subnetwork.primary.id
  gcp_project = var.gcp_project

  instance_type         = var.instance_type
  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  chain      = var.chain
  expose_ssh = "true"

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  region = var.gcp_regions[0]

  instance_count       = var.instance_count[0]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]

  health_check = google_compute_health_check.autohealing.self_link

  gcp_ssh_user    = var.gcp_ssh_user
  gcp_ssh_pub_key = var.gcp_ssh_pub_key

  providers = {
    google = google.primary
  }
  docker_image = var.docker_image
}

module "secondary_region" {
  source = "./modules/regional_infrastructure"

  sa_email = google_service_account.service_account.email

  prefix      = var.prefix
  subnet      = google_compute_subnetwork.secondary.id
  gcp_project = var.gcp_project

  instance_type         = var.instance_type
  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  chain      = var.chain
  expose_ssh = "true"

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  region = var.gcp_regions[1]

  instance_count       = var.instance_count[0]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]

  health_check = google_compute_health_check.autohealing.self_link

  gcp_ssh_user    = var.gcp_ssh_user
  gcp_ssh_pub_key = var.gcp_ssh_pub_key

  providers = {
    google = google.secondary
  }
  docker_image = var.docker_image
}

module "tertiary_region" {
  source = "./modules/regional_infrastructure"

  sa_email = google_service_account.service_account.email

  prefix      = var.prefix
  subnet      = google_compute_subnetwork.tertiary.id
  gcp_project = var.gcp_project

  instance_type         = var.instance_type
  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  chain      = var.chain
  expose_ssh = "true"

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  region = var.gcp_regions[2]

  instance_count       = var.instance_count[0]
  total_instance_count = var.instance_count[0] + var.instance_count[1] + var.instance_count[2]

  health_check = google_compute_health_check.autohealing.self_link

  gcp_ssh_user    = var.gcp_ssh_user
  gcp_ssh_pub_key = var.gcp_ssh_pub_key

  providers = {
    google = google.tertiary
  }
  docker_image = var.docker_image
}
