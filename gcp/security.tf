resource "google_service_account" "service_account" {
  provider = google.primary
  project      = var.gcp_project != "" ? var.gcp_project : null
  account_id   = "${var.prefix}-sa"
  display_name = "SA for Polkadot failover node with prefix ${var.prefix}"
}

resource "google_project_iam_binding" "project" {
  provider = google.primary
  project = var.gcp_project != "" ? var.gcp_project : null

  role    = "roles/compute.viewer"
  members = ["serviceAccount:${google_service_account.service_account.email}"]

}

resource "google_project_iam_binding" "ssm-project" {
  provider = google.primary
  project = var.gcp_project != "" ? var.gcp_project : null

  role    = "roles/secretmanager.viewer"
  members = ["serviceAccount:${google_service_account.service_account.email}"]

}

resource "google_project_iam_binding" "metricswriter-project" {
  provider = google.primary
  project = var.gcp_project != "" ? var.gcp_project : null

  role    = "roles/monitoring.metricWriter"
  members = ["serviceAccount:${google_service_account.service_account.email}"]

}

resource "google_secret_manager_secret_iam_binding" "binding-keys" {
  provider = google.primary

  for_each = var.validator_keys

  project = google_secret_manager_secret.keys[each.key].project
  secret_id = google_secret_manager_secret.keys[each.key].secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}

resource "google_secret_manager_secret_iam_binding" "binding-seeds" {
  provider = google.primary

  for_each = var.validator_keys

  project = google_secret_manager_secret.seeds[each.key].project
  secret_id = google_secret_manager_secret.seeds[each.key].secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}

resource "google_secret_manager_secret_iam_binding" "binding-types" {
  provider = google.primary

  for_each = var.validator_keys

  project = google_secret_manager_secret.types[each.key].project
  secret_id = google_secret_manager_secret.types[each.key].secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}

resource "google_secret_manager_secret_iam_binding" "binding-name" {
  provider = google.primary

  project = google_secret_manager_secret.name.project
  secret_id = google_secret_manager_secret.name.secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}

resource "google_secret_manager_secret_iam_binding" "binding-cpu" {
  provider = google.primary

  project = google_secret_manager_secret.cpu.project
  secret_id = google_secret_manager_secret.cpu.secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}

resource "google_secret_manager_secret_iam_binding" "binding-ram" {
  provider = google.primary

  project = google_secret_manager_secret.ram.project
  secret_id = google_secret_manager_secret.ram.secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}

resource "google_secret_manager_secret_iam_binding" "binding-nodekey" {
  provider = google.primary

  project = google_secret_manager_secret.nodekey.project
  secret_id = google_secret_manager_secret.nodekey.secret_id
  role = "roles/secretmanager.secretAccessor"
  members = ["serviceAccount:${google_service_account.service_account.email}"]
}
