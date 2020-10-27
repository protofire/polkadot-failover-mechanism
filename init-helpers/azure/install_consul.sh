# Function to generate consul health checks to validate node health

function join_by { local IFS="$1"; shift; echo "$*"; }

function create_consul_healthcheck {

  cat <<EOF > "$CONSUL_CONFIG_DIR/healthcheck.json"
{
   "check": {
      "id": "consul_check",
      "name": "Check Consul health 9500",
      "tcp": "localhost:9500",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8600",
      "name": "Check health 8600",
      "tcp": "localhost:8600",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8300",
      "name": "Check health 8300",
      "tcp": "localhost:8300",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8301",
      "name": "Check health 8301",
      "tcp": "localhost:8301",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8302",
      "name": "Check health 8302",
      "tcp": "localhost:8302",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "30333",
      "name": "Check health 30333",
      "tcp": "localhost:30333",
      "interval": "10s",
      "timeout": "1s"
    }
}
EOF

chown "$CONSUL_USER:$CONSUL_USER" "$CONSUL_CONFIG_DIR/healthcheck.json"

}

# Function to install consul binary, configure systemD unit for it,
# and create configuration. 
# based on https://github.com/hashicorp/terraform-google-consul, 
# but Centos support added and supevisorD removed
# parameters to provide:
#  - prefix
#  - total_instance_count

function install_consul {

  set -ex
  # set up parameters to configure consul installation
  echo "Consul version: ${CONSUL_VERSION:=1.7.3}"
  local -r CONSUL_USER=consul 
  local -r CONSUL_DATA_DIR="/var/lib/consul"
  local -r CONSUL_CONFIG_DIR="/opt/consul/config"
  local -r CONSUL_BIN_DIR="/usr/local/bin"
  local -r CONSUL_CONFIG_FILE="config.json"
  local -r CONFIG_PATH="$CONSUL_CONFIG_DIR/$CONSUL_CONFIG_FILE"

  local -r INSTANCE_IP_ADDRESS=$(hostname -I | cut -d ' ' -f1)
  local -r INSTANCE_ID=$(hostname)

  local -r ui="true"
  local -r server="true"

  prefix=$1
  total_instance_count=$2
  lb_primary="$3"
  lb_secondary="$4"
  lb_tertiary="$5"
  cluster_tag_name=$prefix

  lbs=()

  [ -n "${lb_primary}" ] && lbs+=( "\"${lb_primary}\"" )
  [ -n "${lb_secondary}" ] && lbs+=( "\"${lb_secondary}\"" )
  [ -n "${lb_tertiary}" ] && lbs+=( "\"${lb_tertiary}\"" )

  lbs_str=$(join_by ", ", "${lbs[@]}")

  echo "LoadBalancers: ${lbs_str}"

  # Add consul user
  id $CONSUL_USER >& /dev/null || useradd --system $CONSUL_USER

  # Create and chown folders for consul
  mkdir -p $CONSUL_DATA_DIR
  chown -R $CONSUL_USER:$CONSUL_USER $CONSUL_DATA_DIR
  mkdir -p $CONSUL_CONFIG_DIR
  chown -R $CONSUL_USER:$CONSUL_USER $CONSUL_CONFIG_DIR

  # Install consul binary
  curl -o /tmp/consul.zip https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip
  (cd $CONSUL_BIN_DIR ; unzip -u /tmp/consul.zip && rm /tmp/consul.zip)

  chmod +x $CONSUL_BIN_DIR/consul

  # Create and configure systemD target for consul

  cat <<EOF > /etc/systemd/system/consul.service
[Unit]
Description="HashiCorp Consul - A service mesh solution"
Documentation=https://www.consul.io/
Requires=network-online.target
After=network-online.target
ConditionFileNotEmpty=$CONFIG_PATH
[Service]
Type=simple
User=$CONSUL_USER
Group=$CONSUL_USER
ExecStart=$CONSUL_BIN_DIR/consul agent -config-dir $CONSUL_CONFIG_DIR -data-dir $CONSUL_DATA_DIR
ExecReload=$CONSUL_BIN_DIR/consul reload
KillMode=process
Restart=on-failure
TimeoutSec=300s
LimitNOFILE=65536
EOF

# Create config for consul

cat <<EOF > $CONFIG_PATH
{
  "leave_on_terminate": true,
  "datacenter": "${prefix}",
  "advertise_addr": "$INSTANCE_IP_ADDRESS",
  "bind_addr": "0.0.0.0",
  "client_addr": "0.0.0.0",
  "node_name": "$INSTANCE_ID",
  "retry_join": [
    ${lbs_str}
  ],
  "server": $server,
  "ui": $ui,
  "bootstrap_expect": ${total_instance_count}
}
EOF

# Create consul health check
create_consul_healthcheck 

  # SystemD reload and restart

  systemctl daemon-reload
  if systemctl is-enabled --quiet consul; then
    systemctl reenable consul
  else
    systemctl enable consul
  fi

  if systemctl is-active --quiet consul; then
    systemctl restart consul
  else
    systemctl start consul
  fi  

}

