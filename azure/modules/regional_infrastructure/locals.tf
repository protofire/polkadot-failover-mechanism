locals {

  lb_hc = {
    combo = ["Http", "8080", "/verify/checks"]
  }

  prometheus_lb_port = var.expose_prometheus ? { prometheus = [var.prometheus_port, "Tcp", var.prometheus_port] } : {}

  lb_port = merge({
    http     = ["8500", "Tcp", "8500"]
    dns      = ["8600", "Tcp", "8600"]
    rpc      = ["8300", "Tcp", "8300"]
    lan      = ["8301", "Tcp", "8301"]
    wan      = ["8302", "Tcp", "8302"]
    polkadot = ["30333", "Tcp", "30333"]
  }, local.prometheus_lb_port)

  lb_address = {
    primary            = cidrhost(var.subnet_cidrs[0], 10)
    secondary          = cidrhost(var.subnet_cidrs[1], 10)
    tertiary           = cidrhost(var.subnet_cidrs[2], 10)
    private_lb_address = cidrhost(var.subnet_cidr, 10)
  }

  subnet_ports = flatten([
    "8300",
    "8301",
    "8302",
    "8500",
    "8600",
    var.expose_prometheus ? [var.prometheus_port] : []
  ])

  polkadot_prometheus_port = 9615

}
