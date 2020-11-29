locals {
  validator_metric_name = "${var.metric_family}/${var.metric_name}"
  internal_node_ports_tcp = flatten([
    "8500",
    "8600",
    "8300",
    "8301",
    "8302",
    var.expose_prometheus ? [var.prometheus_port] : []
  ])
  internal_node_ports_udp = ["8500", "8600", "8301", "8302"]
  external_node_ports_tcp = flatten([
    "30333",
    var.expose_ssh ? ["22"] : []
  ])
  external_node_ports_udp = ["30333"]
}
