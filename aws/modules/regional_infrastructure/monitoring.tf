resource "aws_cloudwatch_metric_alarm" "validator_count" {

  alarm_name          = "${var.prefix}-polkadot-validator-count"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 1
  treat_missing_data  = "missing"

  alarm_description = "This alert triggers when there are not enough validators."
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_query {
    id          = "e1"
    expression  = "SUM(METRICS())"
    label       = "Validator state"
    return_data = "true"
  }


  dynamic "metric_query" {

    for_each = data.aws_instances.ec2-asg-instance-info.ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "validator_value"
        period      = 60
        stat        = "Sum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = aws_autoscaling_group.polkadot.name
          instance_id = metric_query.value
        }
      }
    }
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns, data.aws_instances.ec2-asg-instance-info]

}

resource "aws_cloudwatch_metric_alarm" "validator_overflow" {

  alarm_name          = "${var.prefix}-polkadot-validator-overflow"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 1
  treat_missing_data  = "missing"

  alarm_description = "CRITICAL! This alert means that there are more than 1 validator at the same time."
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_query {

    id          = "e2"
    expression  = "SUM(METRICS())"
    label       = "Validator state"
    return_data = "true"

  }

  dynamic "metric_query" {

    for_each = data.aws_instances.ec2-asg-instance-info.ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "validator_value"
        period      = 60
        stat        = "Sum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = aws_autoscaling_group.polkadot.name
          instance_id = metric_query.value
        }
      }
    }
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns, data.aws_instances.ec2-asg-instance-info]

}


resource "aws_cloudwatch_metric_alarm" "health_status" {

  alarm_name          = "${var.prefix}-polkadot-health-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 0
  treat_missing_data  = "missing"

  alarm_description = "This metric monitors the polkadot nodes state"
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_query {

    id          = "e3"
    expression  = "SUM(METRICS())"
    label       = "Health state"
    return_data = "true"

  }

  dynamic "metric_query" {

    for_each = data.aws_instances.ec2-asg-instance-info.ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "health_value"
        period      = 60
        stat        = "Maximum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = aws_autoscaling_group.polkadot.name
          instance_id = metric_query.value
        }
      }
    }
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns, data.aws_instances.ec2-asg-instance-info]

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

  metric_query {
    id          = "e4"
    expression  = "SUM(METRICS())"
    label       = "Consul health state"
    return_data = "true"
  }

  dynamic "metric_query" {

    for_each = data.aws_instances.ec2-asg-instance-info.ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "consul_health_checks_critical"
        period      = 60
        stat        = "Maximum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = aws_autoscaling_group.polkadot.name
          instance_id = metric_query.value
          status      = "passing"
          node        = metric_query.value
          check_name  = "Serf Health Status"
          check_id    = "serfHealth"
        }
      }
    }
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns, data.aws_instances.ec2-asg-instance-info]
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

  metric_query {
    id          = "e5"
    expression  = "MAX(METRICS())"
    label       = "Disk used percent"
    return_data = "true"
  }

  dynamic "metric_query" {

    for_each = data.aws_instances.ec2-asg-instance-info.ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "disk_used_percent"
        period      = 60
        stat        = "Maximum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = aws_autoscaling_group.polkadot.name
          instance_id = metric_query.value
          path        = "/data"
          mode        = "rw"
          fstype      = "xfs"
          device      = "nvme1n1"
        }
      }
    }

  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns, data.aws_instances.ec2-asg-instance-info]

}

data "aws_instances" "ec2-asg-instance-info" {

  instance_tags = {
    Name                        = "${var.prefix}-failover-validator"
    prefix                      = var.prefix
    "aws:autoscaling:groupName" = "${var.prefix}-polkadot-validator"
    region                      = var.region
  }

  instance_state_names = ["running"]

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]

}
