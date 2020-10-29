output "asg" {
  value = aws_autoscaling_group.polkadot
}
output "sns" {
  value = aws_sns_topic.sns
}
