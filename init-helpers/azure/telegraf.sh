cat <<EOF >/etc/telegraf/telegraf.conf
[global_tags]
prefix = "$1" # will tag all metrics with dc=us-east-1

# Configuration for telegraf agent
[agent]
  interval = "60s"
  round_interval = true
  hostname = "$2"
  omit_hostname = false

###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################


# # Send aggregate metrics to Azure Monitor
[[outputs.azure_monitor]]
  timeout = "20s"
  namespace_prefix = "$1/"

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

[[inputs.http_listener_v2]]
  service_address = ":12500"
  path = "/telegraf"
  max_body_size = "1MB"
  data_format = "influx"
EOF
