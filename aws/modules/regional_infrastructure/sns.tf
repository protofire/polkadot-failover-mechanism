resource "aws_sns_topic" "sns" {
  name = "${var.prefix}-polkadot-validator"

  tags = {
    prefix = var.prefix
  }
}
# TODO: Change topic subscription
