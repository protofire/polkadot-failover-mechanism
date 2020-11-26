module "primary_region" {

  depends_on = [
    google_secret_manager_secret_version.cpu-version,
    google_secret_manager_secret_version.keys-version,
    google_secret_manager_secret_version.name-version,
    google_secret_manager_secret_version.nodekey-version,
    google_secret_manager_secret_version.ram-version,
    google_secret_manager_secret_version.seeds-version,
    google_secret_manager_secret_version.types-version,
  ]

  source = "./modules/regional_infrastructure"

  sa_email = google_service_account.service_account.email

  prefix      = var.prefix
  subnet      = google_compute_subnetwork.primary.id
  gcp_project = var.gcp_project

  instance_type         = var.instance_type
  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  chain      = var.chain
  expose_ssh = var.expose_ssh

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  region = var.gcp_regions[0]

  instance_count       = polkadot_failover.polkadot.primary_count
  total_instance_count = sum(polkadot_failover.polkadot.failover_instances)

  health_check = google_compute_health_check.autohealing.self_link

  gcp_ssh_user    = var.gcp_ssh_user
  gcp_ssh_pub_key = var.gcp_ssh_pub_key

  prometheus_port   = var.prometheus_port
  expose_prometheus = var.expose_prometheus

  providers = {
    google = google.primary
  }
  docker_image = var.docker_image
}

module "secondary_region" {

  depends_on = [
    google_secret_manager_secret_version.cpu-version,
    google_secret_manager_secret_version.keys-version,
    google_secret_manager_secret_version.name-version,
    google_secret_manager_secret_version.nodekey-version,
    google_secret_manager_secret_version.ram-version,
    google_secret_manager_secret_version.seeds-version,
    google_secret_manager_secret_version.types-version,
  ]

  source = "./modules/regional_infrastructure"

  sa_email = google_service_account.service_account.email

  prefix      = var.prefix
  subnet      = google_compute_subnetwork.secondary.id
  gcp_project = var.gcp_project

  instance_type         = var.instance_type
  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  chain      = var.chain
  expose_ssh = var.expose_ssh

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  region = var.gcp_regions[1]

  instance_count       = polkadot_failover.polkadot.secondary_count
  total_instance_count = sum(polkadot_failover.polkadot.failover_instances)

  health_check = google_compute_health_check.autohealing.self_link

  gcp_ssh_user    = var.gcp_ssh_user
  gcp_ssh_pub_key = var.gcp_ssh_pub_key

  prometheus_port   = var.prometheus_port
  expose_prometheus = var.expose_prometheus

  providers = {
    google = google.secondary
  }
  docker_image = var.docker_image
}

module "tertiary_region" {

  depends_on = [
    google_secret_manager_secret_version.cpu-version,
    google_secret_manager_secret_version.keys-version,
    google_secret_manager_secret_version.name-version,
    google_secret_manager_secret_version.nodekey-version,
    google_secret_manager_secret_version.ram-version,
    google_secret_manager_secret_version.seeds-version,
    google_secret_manager_secret_version.types-version,
  ]

  source = "./modules/regional_infrastructure"

  sa_email = google_service_account.service_account.email

  prefix      = var.prefix
  subnet      = google_compute_subnetwork.tertiary.id
  gcp_project = var.gcp_project

  instance_type         = var.instance_type
  disk_size             = var.disk_size
  delete_on_termination = var.delete_on_termination

  chain      = var.chain
  expose_ssh = var.expose_ssh

  cpu_limit = var.cpu_limit
  ram_limit = var.ram_limit

  region = var.gcp_regions[2]

  instance_count       = polkadot_failover.polkadot.tertiary_count
  total_instance_count = sum(polkadot_failover.polkadot.failover_instances)

  health_check = google_compute_health_check.autohealing.self_link

  gcp_ssh_user    = var.gcp_ssh_user
  gcp_ssh_pub_key = var.gcp_ssh_pub_key

  prometheus_port   = var.prometheus_port
  expose_prometheus = var.expose_prometheus

  providers = {
    google = google.tertiary
  }
  docker_image = var.docker_image
}

module "prometheus" {
  count                    = var.expose_prometheus ? 1 : 0
  source                   = "./modules/prometheus"
  prefix                   = var.prefix
  sa_email                 = google_service_account.service_account.email
  network                  = google_compute_network.vpc_network.name
  subnets                  = [google_compute_subnetwork.primary.name, google_compute_subnetwork.secondary.name, google_compute_subnetwork.tertiary.name]
  gcp_project              = var.gcp_project
  gcp_ssh_user             = var.gcp_ssh_user
  gcp_ssh_pub_key          = var.gcp_ssh_pub_key
  gcp_regions              = var.gcp_regions
  managed_groups           = [module.primary_region.managed_group, module.secondary_region.managed_group, module.tertiary_region.managed_group]
  instance_type            = var.prometheus_instance_type
  expose_ssh               = var.expose_ssh
  instance_count_primary   = polkadot_failover.polkadot.primary_count
  instance_count_secondary = polkadot_failover.polkadot.secondary_count
  instance_count_tertiary  = polkadot_failover.polkadot.tertiary_count
  prometheus_port          = var.prometheus_port

  providers = {
    google = google.primary
  }

}

