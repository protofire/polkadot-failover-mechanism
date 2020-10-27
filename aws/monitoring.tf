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
    label       = "Validator state"
    return_data = "true"
  }

  dynamic "metric_query" {

    for_each = local.instance_ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "validator_value"
        period      = 60
        stat        = "Sum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = "${var.prefix}-polkadot-validator"
          instance_id = metric_query.value
        }
      }
    }
  }

  depends_on = [
    data.aws_instances.ec2-asg-instances-info-1,
    data.aws_instances.ec2-asg-instances-info-2,
    data.aws_instances.ec2-asg-instances-info-3,
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
    label       = "Validator state"
    return_data = "true"

  }

  dynamic "metric_query" {

    for_each = local.instance_ids

    content {
      id = "m${metric_query.key}"
      metric {
        metric_name = "validator_value"
        period      = 60
        stat        = "Sum"
        namespace   = var.prefix
        dimensions = {
          asg_name    = "${var.prefix}-polkadot-validator"
          instance_id = metric_query.value
        }
      }
    }
  }

  depends_on = [
    data.aws_instances.ec2-asg-instances-info-1,
    data.aws_instances.ec2-asg-instances-info-2,
    data.aws_instances.ec2-asg-instances-info-3,
    module.primary_region.sns,
  ]
}

data "aws_instances" "ec2-asg-instances-info-1" {

  provider = aws.primary

  instance_tags = {
    Name                        = "${var.prefix}-failover-validator"
    prefix                      = var.prefix
    "aws:autoscaling:groupName" = "${var.prefix}-polkadot-validator"
  }

  instance_state_names = ["running"]

  depends_on = [module.primary_region.asg]

}

data "aws_instances" "ec2-asg-instances-info-2" {

  provider = aws.secondary

  instance_tags = {
    Name                        = "${var.prefix}-failover-validator"
    prefix                      = var.prefix
    "aws:autoscaling:groupName" = "${var.prefix}-polkadot-validator"
  }

  instance_state_names = ["running"]

  depends_on = [module.secondary_region.asg]

}

data "aws_instances" "ec2-asg-instances-info-3" {

  provider = aws.tertiary

  instance_tags = {
    Name                        = "${var.prefix}-failover-validator"
    prefix                      = var.prefix
    "aws:autoscaling:groupName" = "${var.prefix}-polkadot-validator"
  }

  instance_state_names = ["running"]

  depends_on = [module.tertiary_region.asg]

}