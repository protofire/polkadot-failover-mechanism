resource "aws_cloudwatch_metric_alarm" "health_status" {

  alarm_name          = "${var.prefix}-polkadot-health-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 0
  treat_missing_data  = "missing"

  alarm_description = "This metric monitors the polkadot nodes state"
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "health_value"
  period      = 60
  statistic   = "Maximum"
  namespace   = var.prefix
  dimensions = {
    group_name = aws_autoscaling_group.polkadot.name
  }
  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]

}

resource "aws_cloudwatch_metric_alarm" "consul_health_status" {

  alarm_name          = "${var.prefix}-polkadot-consul-health-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 0
  treat_missing_data  = "missing"

  alarm_description = "This metric monitors the polkadot nodes state"
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "consul_health_checks_critical"
  period      = 60
  statistic   = "Maximum"
  namespace   = var.prefix
  dimensions = {
    group_name = aws_autoscaling_group.polkadot.name
    status     = "passing"
    check_name = "Serf Health Status"
    check_id   = "serfHealth"
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]

}

resource "aws_cloudwatch_metric_alarm" "disk_used_percent" {

  alarm_name          = "${var.prefix}-polkadot-disk-used-percent"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 90
  treat_missing_data  = "missing"

  alarm_description = "CRITICAL! This alert means that the instance disk is running out of space"
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "disk_used_percent"
  period      = 60
  statistic   = "Maximum"
  namespace   = var.prefix
  dimensions = {
    group_name = aws_autoscaling_group.polkadot.name
    path       = "/data"
    mode       = "rw"
    fstype     = "xfs"
    device     = "nvme1n1"
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]

}
