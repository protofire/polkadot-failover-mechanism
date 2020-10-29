# Helper function to get IP address of consul instance from metadata server
function get_instance_ip_address {
  local network_interface_number="$1"

  readonly COMPUTE_INSTANCE_METADATA_URL="http://metadata.google.internal/computeMetadata/v1"
  readonly GOOGLE_CLOUD_METADATA_REQUEST_HEADER="Metadata-Flavor: Google"

  # If no network interface number was specified, default to the first one
  if [[ -z "$network_interface_number" ]]; then
    network_interface_number=0
  fi

  # echo "Looking up Compute Instance IP Address on Network Interface $network_interface_number"
  curl --silent  --header "$GOOGLE_CLOUD_METADATA_REQUEST_HEADER" \
  "$COMPUTE_INSTANCE_METADATA_URL/instance/network-interfaces/${network_interface_number}/ip"
}

# Helper function to get instance name of consul instance from metadata server
function get_instance_name {
  
  readonly COMPUTE_INSTANCE_METADATA_URL="http://metadata.google.internal/computeMetadata/v1"
  readonly GOOGLE_CLOUD_METADATA_REQUEST_HEADER="Metadata-Flavor: Google"
  #echo "Looking up Compute Instance name"
  curl --silent  --header "$GOOGLE_CLOUD_METADATA_REQUEST_HEADER" \
  "$COMPUTE_INSTANCE_METADATA_URL/instance/name"  
}

# Helper function to get project ID from metadata server
function get_project_id {

  readonly COMPUTE_INSTANCE_METADATA_URL="http://metadata.google.internal/computeMetadata/v1"
  readonly GOOGLE_CLOUD_METADATA_REQUEST_HEADER="Metadata-Flavor: Google"
  # echo "Looking up Project ID"
  curl --silent  --header "$GOOGLE_CLOUD_METADATA_REQUEST_HEADER" \
  "$COMPUTE_INSTANCE_METADATA_URL/project/project-id"  
}

# Function to generate consul health checks to validate node health
function create_consul_healthcheck {

  cat <<EOF > "${CONSUL_CONFIG_DIR}/healthcheck.json"
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

chown "${CONSUL_USER}:${CONSUL_USER}" "${CONSUL_CONFIG_DIR}/healthcheck.json"

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
  echo "Consul version: ${CONSUL_VERSION:=1.7.2}"
  local -r CONSUL_USER=consul 
  local -r CONSUL_DATA_DIR="/var/lib/consul"
  local -r CONSUL_CONFIG_DIR="/opt/consul/config"
  local -r CONSUL_BIN_DIR="/usr/local/bin"
  local -r CONSUL_CONFIG_FILE="config.json"
  local -r CONFIG_PATH="$CONSUL_CONFIG_DIR/$CONSUL_CONFIG_FILE"

  local -r INSTANCE_IP_ADDRESS=$(get_instance_ip_address 0)
  local -r INSTANCE_ID=$(get_instance_name)
  local -r PROJECT_ID=$(get_project_id)

  local -r ui="true"
  local -r server="true"

  prefix=$1
  total_instance_count=$2
  cluster_tag_name=$prefix

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
  "bind_addr": "$INSTANCE_IP_ADDRESS",
  "client_addr": "0.0.0.0",
  "node_name": "$INSTANCE_ID",
  "retry_join": ["provider=gce project_name=$PROJECT_ID tag_value=$cluster_tag_name"],
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
