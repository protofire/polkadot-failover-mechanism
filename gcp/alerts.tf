resource "google_monitoring_notification_channel" "polkadot" {
  provider = google.primary
  project  = var.gcp_project != "" ? var.gcp_project : null

  display_name = "${var.prefix}-notifications"
  type         = "email"

  labels = {
    email_address = var.admin_email
  }
  depends_on = [module.primary_region, module.secondary_region, module.tertiary_region]
}

resource "google_monitoring_alert_policy" "validator" {
  provider = google.primary
  project  = var.gcp_project != "" ? var.gcp_project : null

  display_name = "${var.prefix}-validator-min"
  combiner     = "OR"

  conditions {
    display_name = "Health not OK"
    condition_threshold {
      filter          = "metric.type=\"custom.googleapis.com/${var.metrics_namespace}/health/value\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration        = "60s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0
      trigger {
        count = 1
      }
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  conditions {
    display_name = "Health not OK"
    condition_threshold {
      filter          = "metric.type=\"custom.googleapis.com/${var.metrics_namespace}/health/value\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration        = "60s"
      comparison      = "COMPARISON_LT"
      threshold_value = 0
      trigger {
        count = 1
      }
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MIN"
      }
    }
  }

  conditions {
    display_name = "Validator less than 1"
    condition_threshold {
      filter          = "metric.type=\"custom.googleapis.com/${var.metrics_namespace}/validator/value\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration        = "60s"
      comparison      = "COMPARISON_LT"
      threshold_value = 1
      trigger {
        count = 1
      }
      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MIN"
        cross_series_reducer = "REDUCE_MAX"
      }
    }
  }

  conditions {
    display_name = "Validator more than 1"
    condition_threshold {
      filter          = "metric.type=\"custom.googleapis.com/${var.metrics_namespace}/validator/value\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration        = "60s"
      comparison      = "COMPARISON_GT"
      threshold_value = 1
      trigger {
        count = 1
      }
      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MAX"
        cross_series_reducer = "REDUCE_SUM"
      }
    }
  }

  conditions {
    display_name = "Disk consumed more than 90 percent"
    condition_threshold {
      filter          = "metric.type=\"custom.googleapis.com/${var.metrics_namespace}/disk/used_percent\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration        = "60s"
      comparison      = "COMPARISON_GT"
      threshold_value = 90
      trigger {
        count = 1
      }
      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MAX"
        cross_series_reducer = "REDUCE_SUM"
      }
    }
  }

  user_labels = {
    prefix = var.prefix
  }

  notification_channels = [google_monitoring_notification_channel.polkadot.name]

}
