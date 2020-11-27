prefix=$1
hostname=$2
group_name=$3
expose_prometheus=${4:-false}
polkadot_prometheus_port=$5
prometheus_port=$6

cat <<EOF >/etc/telegraf/telegraf.conf
[global_tags]
  prefix = "${prefix}" # will tag all metrics with dc=us-east-1
  instance_id = "${hostname}"
  group_name = "${group_name}"

# Configuration for telegraf agent
[agent]
  interval = "60s"
  round_interval = true
  hostname = "${hostname}"
  omit_hostname = false

###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################


# # Send aggregate metrics to Azure Monitor
[[outputs.azure_monitor]]
  timeout = "20s"
  namespace_prefix = "${prefix}/"
  strings_as_dimensions = true
EOF

if [ "${expose_prometheus}" = true ]; then
cat <<EOF >>/etc/telegraf/telegraf.conf

[[outputs.prometheus_client]]
  listen = ":${prometheus_port}"
  metric_version = 2
  expiration_interval = "120s"
  path = "/metrics"
  collectors_exclude = ["gocollector", "process"]
EOF
fi

cat <<EOF >>/etc/telegraf/telegraf.conf

###############################################################################
#                            INPUT PLUGINS                                    #
###############################################################################

# Read metrics about cpu usage
[[inputs.cpu]]
  totalcpu = true

# Read metrics about disk usage by mount point
[[inputs.disk]]
  mount_points = ["/data", "/"]
  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

# Read metrics about memory usage
[[inputs.mem]]
  # no configuration

# Read metrics about swap memory usage
[[inputs.swap]]
  # no configuration

[[inputs.consul]]
  datacenter = "${prefix}"
  metric_version = 2

[[inputs.http_listener_v2]]
  service_address = ":12500"
  path = "/telegraf"
  max_body_size = "1MB"
  data_format = "influx"
EOF

if [ "${expose_prometheus}" = true ]; then
cat <<EOF >>/etc/telegraf/telegraf.conf

[[inputs.prometheus]]
  urls = ["http://localhost:${polkadot_prometheus_port}/metrics"]
  metric_version = 2
  response_timeout = "5s"
  interval = "10s"
  [inputs.prometheus.tags]
    job = "polkadot"
EOF
fi
