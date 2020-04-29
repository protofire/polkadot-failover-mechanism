resource "aws_cloudwatch_metric_alarm" "validator_overflow" {
  alarm_name          = "${var.prefix}-polkadot-validator-overflow"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 1
  treat_missing_data  = "breaching"

  alarm_description = "This alert triggers when there are not enough validatrs."
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "Validator count"
  namespace   = var.prefix
  period      = 60
  statistic   = "Sum"

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.polkadot.name
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]
}

resource "aws_cloudwatch_metric_alarm" "validator_count" {
  alarm_name          = "${var.prefix}-polkadot-validator-count"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 1
  treat_missing_data  = "breaching"

  alarm_description = "CRITICAL! This alert means that there are more than 1 validator at the same time."
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "Validator count"
  namespace   = var.prefix
  period      = 60
  statistic   = "Sum"

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.polkadot.name
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]
}


resource "aws_cloudwatch_metric_alarm" "count" {
  alarm_name          = "${var.prefix}-polkadot-node-count"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = var.total_instance_count
  treat_missing_data  = "breaching"

  alarm_description = "This metric monitors the polkadot nodes state"
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "Health report"
  namespace   = var.prefix
  period      = 60
  statistic   = "SampleCount"

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.polkadot.name
  }

  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]
}

resource "aws_cloudwatch_metric_alarm" "status" {
  alarm_name          = "${var.prefix}-polkadot-failover-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = "0"
  treat_missing_data  = "breaching"

  alarm_description = "This metric monitors the polkadot nodes state"
  alarm_actions     = [aws_sns_topic.sns.arn]

  metric_name = "Health report"
  namespace   = var.prefix
  period      = 60
  statistic   = "Maximum"

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.polkadot.name
  }


  depends_on = [aws_autoscaling_group.polkadot, aws_sns_topic.sns]
}
