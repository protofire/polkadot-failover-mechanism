resource "aws_lb_listener" "http" {
  load_balancer_arn = var.lb.arn
  port              = "8500"
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.http.arn
  }
}

resource "aws_lb_listener" "dns" {
  load_balancer_arn = var.lb.arn
  port              = "8600"
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.dns.arn
  }
}

resource "aws_lb_listener" "rpc" {
  load_balancer_arn = var.lb.arn
  port              = "8300"
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.rpc.arn
  }
}

resource "aws_lb_listener" "lan" {
  load_balancer_arn = var.lb.arn
  port              = "8301"
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.lan.arn
  }
}

resource "aws_lb_listener" "wan" {
  load_balancer_arn = var.lb.arn
  port              = "8302"
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.wan.arn
  }
}

resource "aws_lb_listener" "polkadot" {
  load_balancer_arn = var.lb.arn
  port              = "30333"
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.polkadot.arn
  }
}

resource "aws_lb_listener" "prometheus" {
  count             = var.expose_prometheus ? 1 : 0
  load_balancer_arn = var.lb.arn
  port              = var.prometheus_port
  protocol          = "TCP_UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.prometheus[count.index].arn
  }
}
