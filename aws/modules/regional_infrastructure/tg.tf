resource "aws_lb_target_group" "http" {
  name     = "${var.prefix}-polkadot-validator-http"
  port     = 8500
  protocol = "TCP_UDP"
  vpc_id   = var.vpc.id

  deregistration_delay = 1

  health_check {

    enabled             = true
    protocol            = "TCP"
    interval            = var.health_check_interval
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold

  }
}

resource "aws_lb_target_group" "dns" {
  name     = "${var.prefix}-polkadot-validator-dns"
  port     = 8600
  protocol = "TCP_UDP"
  vpc_id   = var.vpc.id

  deregistration_delay = 1

  health_check {

    enabled             = true
    protocol            = "TCP"
    interval            = var.health_check_interval
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold

  }
}

resource "aws_lb_target_group" "rpc" {
  name     = "${var.prefix}-polkadot-validator-rpc"
  port     = 8300
  protocol = "TCP_UDP"
  vpc_id   = var.vpc.id

  deregistration_delay = 1

  health_check {

    enabled             = true
    protocol            = "TCP"
    interval            = var.health_check_interval
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold

  }
}

resource "aws_lb_target_group" "lan" {
  name     = "${var.prefix}-polkadot-validator-lan"
  port     = 8301
  protocol = "TCP_UDP"
  vpc_id   = var.vpc.id

  deregistration_delay = 1

  health_check {

    enabled             = true
    protocol            = "TCP"
    interval            = var.health_check_interval
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold

  }
}

resource "aws_lb_target_group" "wan" {
  name     = "${var.prefix}-polkadot-validator-wan"
  port     = 8302
  protocol = "TCP_UDP"
  vpc_id   = var.vpc.id

  deregistration_delay = 1

  health_check {

    enabled             = true
    protocol            = "TCP"
    interval            = var.health_check_interval
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold

  }
}

resource "aws_lb_target_group" "polkadot" {
  name     = "${var.prefix}-polkadot-validator"
  port     = 30333
  protocol = "TCP_UDP"
  vpc_id   = var.vpc.id

  deregistration_delay = 1

  health_check {

    enabled             = true
    protocol            = "TCP"
    interval            = var.health_check_interval
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold

  }
}
