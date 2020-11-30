resource "aws_security_group" "validator-node" {
  name        = "${var.prefix}-polkadot-validator"
  description = "A security group for the app server ALB, so it is accessible via the web"
  vpc_id      = var.vpc.id

  dynamic "ingress" {
    for_each = var.expose_ssh ? [1] : []
    content {
      from_port   = 22
      to_port     = 22
      protocol    = "TCP"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.expose_prometheus ? [1] : []
    content {
      from_port   = var.prometheus_port
      to_port     = var.prometheus_port
      protocol    = "TCP"
      cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
    }
  }

  ingress {
    from_port   = 30333
    to_port     = 30333
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 30333
    to_port     = 30333
    protocol    = "UDP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8500
    to_port     = 8500
    protocol    = "TCP"
    cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
  }

  ingress {
    from_port   = 8500
    to_port     = 8500
    protocol    = "UDP"
    cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
  }

  ingress {
    from_port   = 8600
    to_port     = 8600
    protocol    = "TCP"
    cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
  }

  ingress {
    from_port   = 8600
    to_port     = 8600
    protocol    = "UDP"
    cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
  }

  ingress {
    from_port   = 8300
    to_port     = 8302
    protocol    = "TCP"
    cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
  }

  ingress {
    from_port   = 8301
    to_port     = 8302
    protocol    = "UDP"
    cidr_blocks = [var.cidrs[0], var.cidrs[1], var.cidrs[2]]
  }

  # Unrestricted outbound
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
