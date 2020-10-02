resource "aws_lb" "polkadot" {
  name                             = "${var.prefix}-internal-lb-polkadot"
  internal                         = true
  load_balancer_type               = "network"
  enable_cross_zone_load_balancing = true
  subnets                          = [aws_subnet.default.id]

  tags = {
    prefix = var.prefix
  }
}
