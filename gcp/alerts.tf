resource "google_monitoring_notification_channel" "polkadot" {
  display_name = "${var.prefix}-notifications"
  type         = "email"

  labels = {
    email_address = var.admin_email
  }
}

resource "google_monitoring_alert_policy" "validator" {
  display_name = "${var.prefix}-validator-min"
  combiner     = "OR"

  conditions {
    display_name = "Health not OK"
    condition_threshold {
      filter     = "metric.type=\"custom.googleapis.com/polkadot/polkadot/state_\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration   = "60s"
      comparison = "COMPARISON_GT"
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
      filter     = "metric.type=\"custom.googleapis.com/polkadot/polkadot/state_\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration   = "60s"
      comparison = "COMPARISON_LT"
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
      filter     = "metric.type=\"custom.googleapis.com/polkadot/polkadot/validatorcount_\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration   = "60s"
      comparison = "COMPARISON_LT"
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
      filter     = "metric.type=\"custom.googleapis.com/polkadot/polkadot/validatorcount_\" AND resource.type=\"gce_instance\" AND metadata.user_labels.prefix=\"${var.prefix}\""
      duration   = "60s"
      comparison = "COMPARISON_GT"
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

  user_labels = {
    prefix = var.prefix
  }

  notification_channels = [google_monitoring_notification_channel.polkadot.name]

  depends_on = [null_resource.delay]
}

resource "null_resource" "delay" {
  provisioner "local-exec" {
    command = "sleep 400"
  }
  triggers = {
    before = google_monitoring_notification_channel.polkadot.name 
  }
}
