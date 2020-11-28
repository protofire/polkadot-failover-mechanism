namespace=$1
hostname=$2
group_name=$3
instance_id=$4
region1=$5
region2=$6
region3=$7
expose_prometheus=${8:-false}
polkadot_prometheus_port=$9
prometheus_port=${10}

cat <<EOF >/etc/telegraf/telegraf.conf
[global_tags]
  prefix = "${namespace}"
  group_name = "${group_name}"

[agent]
  interval = "60s"
  round_interval = true
  hostname = "${hostname}"
  omit_hostname = true

###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

# Send aggregate metrics to AWS cloudwatch
[[outputs.cloudwatch]]
  region = "${region1}"
  namespace = "${namespace}"
  high_resolution_metrics = false
  write_statistics = false
  namedrop = ["*prometheus*"]
  tagexclude = ["prefix", "job"]

[[outputs.cloudwatch]]
  region = "${region2}"
  namespace = "${namespace}"
  high_resolution_metrics = false
  write_statistics = false
  namedrop = ["*prometheus*"]
  tagexclude = ["prefix", "job"]

[[outputs.cloudwatch]]
  region = "${region3}"
  namespace = "${namespace}"
  high_resolution_metrics = false
  write_statistics = false
  namedrop = ["*prometheus*"]
  tagexclude = ["prefix", "job"]
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
  datacenter = "${namespace}"
  metric_version = 2
  tagexclude = ["node"]

[[inputs.http_listener_v2]]
  service_address = ":12500"
  path = "/telegraf"
  max_body_size = "1MB"
  data_format = "influx"

[[inputs.http_listener_v2]]
  service_address = ":12501"
  path = "/telegraf"
  max_body_size = "1MB"
  data_format = "influx"
  [inputs.http_listener_v2.tags]
    instance_id = "${instance_id}"
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
    instance_id = "${instance_id}"
    group_name = "${group_name}"
EOF
fi
