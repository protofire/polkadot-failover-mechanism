locals {
  polkadot_prometheus_port = 9615
  group_name               = "${var.prefix}-instance-group-manager"
  group_name_region        = "${var.prefix}-instance-group-manager-${var.region}"
}
