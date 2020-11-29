resource "aws_security_group" "prometheus" {
  name        = "${var.prefix}-polkadot-prometheus"
  description = "A security group for the app server ALB, so it is accessible via the web"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.expose_ssh ? [1] : []
    content {
      from_port   = 22
      to_port     = 22
      protocol    = "TCP"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  ingress {
    from_port   = var.prometheus_port
    to_port     = var.prometheus_port
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    prefix = var.prefix
  }
}
