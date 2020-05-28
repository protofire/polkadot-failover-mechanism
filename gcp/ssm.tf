resource "google_secret_manager_secret_version" "keys-version" {
  provider = google-beta

  for_each = var.validator_keys

  secret = google_secret_manager_secret.keys[each.key].id

  secret_data = each.value.key
}

resource "google_secret_manager_secret" "keys" {
  provider = google-beta

  for_each = var.validator_keys

  labels = {
    prefix = var.prefix
    type = "key"
  }

  secret_id        = "${var.prefix}_${each.key}_key"
  
  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}

resource "google_secret_manager_secret_version" "seeds-version" {
  provider = google-beta

  for_each = var.validator_keys

  secret = google_secret_manager_secret.seeds[each.key].id

  secret_data = each.value.seed
}

resource "google_secret_manager_secret" "seeds" {
  provider = google-beta

  for_each = var.validator_keys

  labels = {
    prefix = var.prefix
    type = "key"
  }

  secret_id        = "${var.prefix}_${each.key}_seed"

  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}

resource "google_secret_manager_secret_version" "types-version" {
  provider = google-beta

  for_each = var.validator_keys

  secret = google_secret_manager_secret.types[each.key].id

  secret_data = each.value.type
}

resource "google_secret_manager_secret" "types" {
  provider = google-beta

  for_each = var.validator_keys

  labels = {
    prefix = var.prefix
    type = "key"
  }

  secret_id        = "${var.prefix}_${each.key}_type"

  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}

resource "google_secret_manager_secret_version" "name-version" {
  provider = google-beta

  secret = google_secret_manager_secret.name.id

  secret_data = var.validator_name 
}

resource "google_secret_manager_secret" "name" {
  provider = google-beta

  secret_id = "${var.prefix}_name"

  labels = {
    prefix = var.prefix
  }

  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}

resource "google_secret_manager_secret_version" "cpu-version" {
  provider = google-beta

  secret = google_secret_manager_secret.cpu.id

  secret_data = var.cpu_limit
}

resource "google_secret_manager_secret" "cpu" {
  provider = google-beta

  secret_id = "${var.prefix}_cpulimit"

  labels = {
    prefix = var.prefix
  }

  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}

resource "google_secret_manager_secret_version" "ram-version" {
  provider = google-beta

  secret = google_secret_manager_secret.ram.id

  secret_data = var.ram_limit
}

resource "google_secret_manager_secret" "ram" {
  provider = google-beta

  secret_id = "${var.prefix}_ramlimit"

  labels = {
    prefix = var.prefix
  }

  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}

resource "google_secret_manager_secret_version" "nodekey-version" {
  provider = google-beta

  secret = google_secret_manager_secret.nodekey.id

  secret_data = var.node_key
}

resource "google_secret_manager_secret" "nodekey" {
  provider = google-beta

  secret_id = "${var.prefix}_nodekey"

  labels = {
    prefix = var.prefix
  }

  replication {
    user_managed {
      replicas {
        location = var.gcp_regions[0]
      }
      replicas {
        location = var.gcp_regions[1]
      }
      replicas {
        location = var.gcp_regions[2]
      }
    }
  }
}
