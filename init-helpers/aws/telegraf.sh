cat <<EOF >/etc/telegraf/telegraf.conf
[global_tags]
asg_name = "$4" # will tag all metrics with asg name
instance_id = "$5" # will tag all metrics with instance id

# Configuration for telegraf agent
[agent]
  interval = "60s"
  round_interval = true
  hostname = "$2"
  omit_hostname = true

###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

# Send aggregate metrics to AWS cloudwatch
[[outputs.cloudwatch]]
  region = "$3"
  namespace = "$1"
  high_resolution_metrics = false
  write_statistics = false

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
datacenter = "$1"
metric_version = 2

[[inputs.http_listener_v2]]
  service_address = ":12500"
  path = "/telegraf"
  max_body_size = "1MB"
  data_format = "influx"
EOF
