#!/usr/bin/env bash

# Set quit on error flags
set -x -eE

# Install unzip, docker, jq, git, awscli
curl -o /etc/yum.repos.d/influxdb.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/influxdb.repo
/usr/bin/yum update -y
/usr/bin/yum install unzip jq git telegraf nc -y

urls=()
function join_by { local IFS="$1"; shift; echo "$*"; }

[ -n "${url_primary}" ] && urls+=( "\"${url_primary}\"" )
[ -n "${url_secondary}" ] && urls+=( "\"${url_secondary}\"" )
[ -n "${url_tertiary}" ] && urls+=( "\"${url_tertiary}\"" )

urls_str=$(join_by ", ", "$${urls[@]}")
echo "LoadBalancers: $urls_str"

cat <<EOF >/etc/telegraf/telegraf.conf
[global_tags]

# Configuration for telegraf agent
[agent]
  interval = "60s"
  round_interval = true
  omit_hostname = true

###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

[[outputs.prometheus_client]]
  listen = ":${prometheus_port}"
  metric_version = 2
  expiration_interval = "120s"
  path = "/metrics"
  collectors_exclude = ["gocollector", "process"]

###############################################################################
#                            INPUT PLUGINS                                    #
###############################################################################

[[inputs.prometheus]]
  urls = [$urls_str]
  metric_version = 2
  response_timeout = "5s"
  interval = "10s"
EOF

/usr/bin/systemctl enable telegraf
/usr/bin/systemctl restart telegraf
