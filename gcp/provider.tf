provider "google-beta" {
  version = "~> 3.15.0"
  credentials = var.gcp_credentials != "" ? file(var.gcp_credentials) : null
  project = var.gcp_project != "" ? var.gcp_project : null
  region  = var.gcp_regions[0]
}

provider "google" {
  version = "~> 3.15.0"
  credentials = var.gcp_credentials != "" ? file(var.gcp_credentials) : null
  project = var.gcp_project != "" ? var.gcp_project : null
  region  = var.gcp_regions[0]
  alias = "primary" 
}

provider "google" {
  version = "~> 3.15.0"
  credentials = var.gcp_credentials != "" ? file(var.gcp_credentials) : null
  project = var.gcp_project != "" ? var.gcp_project : null
  region  = var.gcp_regions[1]
  alias = "secondary"
}

provider "google" {
  version = "~> 3.15.0"
  credentials = var.gcp_credentials != "" ? file(var.gcp_credentials) : null
  project = var.gcp_project != "" ? var.gcp_project : null
  region  = var.gcp_regions[2]
  alias = "tertiary"
}
