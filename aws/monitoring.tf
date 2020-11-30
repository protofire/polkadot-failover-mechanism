resource "aws_cloudwatch_metric_alarm" "validator_count" {

  provider = aws.primary

  alarm_name          = "${var.prefix}-polkadot-validator-count"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 1
  treat_missing_data  = "missing"

  alarm_description = "This alert triggers when there are not enough validators."
  alarm_actions     = [module.primary_region.sns.arn]
  ok_actions        = [module.primary_region.sns.arn]

  metric_query {
    id          = "e1"
    expression  = "SUM(METRICS())"
    label       = "Validators count"
    return_data = "true"
  }

  dynamic "metric_query" {

    for_each = [
      module.primary_region.asg.name,
      module.secondary_region.asg.name,
    module.tertiary_region.asg.name]

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "validator_value"
        period      = 60
        stat        = "Maximum"
        namespace   = var.prefix
        dimensions = {
          group_name = metric_query.value
        }
      }
    }
  }

  depends_on = [
    module.primary_region.asg,
    module.secondary_region.asg,
    module.tertiary_region.asg,
    module.primary_region.sns,
  ]

}


resource "aws_cloudwatch_metric_alarm" "validator_overflow" {

  provider = aws.primary

  alarm_name          = "${var.prefix}-polkadot-validator-overflow"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  datapoints_to_alarm = 3
  threshold           = 1
  treat_missing_data  = "missing"

  alarm_description = "CRITICAL! This alert means that there are more than 1 validator at the same time."
  alarm_actions     = [module.primary_region.sns.arn]
  ok_actions        = [module.primary_region.sns.arn]

  metric_query {
    id          = "e2"
    expression  = "SUM(METRICS())"
    label       = "Validators overflow"
    return_data = "true"
  }

  dynamic "metric_query" {

    for_each = [
      module.primary_region.asg.name,
      module.secondary_region.asg.name,
    module.tertiary_region.asg.name]

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "validator_value"
        period      = 60
        stat        = "Maximum"
        namespace   = var.prefix
        dimensions = {
          group_name = metric_query.value
        }
      }
    }
  }

  depends_on = [
    module.primary_region.asg,
    module.secondary_region.asg,
    module.tertiary_region.asg,
    module.primary_region.sns,
  ]
}
